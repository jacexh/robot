package robot

import (
	"errors"
	"log/slog"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"gocv.io/x/gocv"
)

type Camera struct {
	ID      int
	output  chan gocv.Mat
	fanouts map[string]chan gocv.Mat
	mu      sync.RWMutex
}

type Receiver interface {
}

func NewCamera() *Camera {
	return &Camera{ID: 0, output: make(chan gocv.Mat, 1), fanouts: make(map[string]chan gocv.Mat)}
}

func (c *Camera) Start() error {
	cam, err := gocv.OpenVideoCapture(c.ID)
	if err != nil {
		return err
	}

	go func() {
		slog.Info("Starting camera", "id", c.ID)
		c.mu.RLock()
		subs := make([]chan gocv.Mat, 0, len(c.fanouts))
		for _, sub := range c.fanouts {
			subs = append(subs, sub)
		}
		c.mu.RUnlock()

		for img := range c.output {
			for _, s := range subs {
				s <- img.Clone()
			}
		}
	}()

	defer cam.Close()

	img := gocv.NewMat()
	defer img.Close()

	for {
		if ok := cam.Read(&img); !ok {
			return errors.New("failed to read from camera")
		}
		if img.Empty() {
			continue
		}
	}
}

func (c *Camera) Stream(w http.ResponseWriter, r *http.Request) {
	receiver := make(chan gocv.Mat, 1)
	id := uuid.New().String()
	c.mu.Lock()
	c.fanouts[id] = receiver
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.fanouts, id)
		c.mu.Unlock()
		close(receiver)
	}()

	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")

	for img := range receiver {
		buf, err := gocv.IMEncode(".jpg", img)
		if err != nil {
			http.Error(w, "Failed to encode image: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte("--frame\r\n"))
		w.Write([]byte("Content-Type: image/jpeg\r\n\r\n"))
		w.Write(buf.GetBytes())
		w.Write([]byte("\r\n"))
	}
}
