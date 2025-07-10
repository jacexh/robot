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
	defer close(c.output)
	cam, err := gocv.OpenVideoCapture(c.ID)
	if err != nil {
		return err
	}

	go func() {
		slog.Info("Starting camera", "id", c.ID)

		for img := range c.output {
			c.mu.RLock()
			pipes := make([]chan gocv.Mat, 0, len(c.fanouts))
			for _, pipe := range c.fanouts {
				pipes = append(pipes, pipe)
			}
			c.mu.RUnlock()

			for _, s := range pipes {
				s <- img
			}
			slog.Info("sent frame")
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
		c.output <- img.Clone()
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
