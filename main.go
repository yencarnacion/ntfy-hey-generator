// main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

type config struct {
	Server string
	Port   string
	Topics []string
	MP3    string
}

func loadConfig() (config, error) {
	if err := godotenv.Load(); err != nil {
		return config{}, fmt.Errorf("failed to load .env: %w", err)
	}

	get := func(key, fallback string) string {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
		return fallback
	}

	cfg := config{
		Server: get("NTFY_SERVER_URL", "ntfy.sh"),
		Port:   get("NTFY_PORT", "80"),
		Topics: strings.Split(get("NTFY_TOPICS", ""), ","),
		MP3:    get("MP3_FILE", ""),
	}

	if len(cfg.Topics) == 0 || cfg.MP3 == "" {
		return cfg, fmt.Errorf("NTFY_TOPICS and MP3_FILE must be defined")
	}
	for i := range cfg.Topics {
		cfg.Topics[i] = strings.TrimSpace(cfg.Topics[i])
	}
	return cfg, nil
}

func prepareAudio(path string) (*beep.Buffer, beep.Format, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, beep.Format{}, fmt.Errorf("open mp3: %w", err)
	}
	defer f.Close()

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return nil, format, fmt.Errorf("decode mp3: %w", err)
	}
	defer streamer.Close()

	buf := beep.NewBuffer(format)
	buf.Append(streamer)
	return buf, format, nil
}

// connect dials the WebSocket, plays on each message, and returns immediately when ctx is cancelled.
func connect(ctx context.Context, u url.URL, play func()) {
	for {
		// abort if context already cancelled
		if ctx.Err() != nil {
			return
		}

		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Printf("dial %s: %v (retrying in 5 s)", u.String(), err)
			select {
			case <-time.After(5 * time.Second):
				continue
			case <-ctx.Done():
				return
			}
		}
		log.Printf("listening on %s", u.String())

		done := make(chan struct{})
		// read loop in its own goroutine
		go func() {
			defer close(done)
			for {
				_, _, err := c.ReadMessage()
				if err != nil {
					return // connection closed or errored
				}
				play()
			}
		}()

		// wait until either context is cancelled or reader exits
		select {
		case <-ctx.Done():
			_ = c.Close() // unblocks ReadMessage
			<-done        // ensure reader goroutine ends
			return        // stop reconnect loop
		case <-done:
			_ = c.Close()
			// reader ended (server closed); reconnect unless ctx cancelled
		}
	}
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	buffer, format, err := prepareAudio(cfg.MP3)
	if err != nil {
		log.Fatal(err)
	}

	// 100 ms audio buffer
	if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
		log.Fatal(err)
	}

	var mixer beep.Mixer
	speaker.Play(&mixer)

	playClip := func() {
		speaker.Lock()
		mixer.Add(buffer.Streamer(0, buffer.Len()))
		speaker.Unlock()
	}

	// handle Ctrl-C / SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	for _, topic := range cfg.Topics {
		if topic == "" {
			continue
		}
		wg.Add(1)
		go func(tp string) {
			defer wg.Done()
			u := url.URL{
				Scheme: "ws",
				Host:   fmt.Sprintf("%s:%s", cfg.Server, cfg.Port),
				Path:   fmt.Sprintf("/%s/ws", tp),
			}
			connect(ctx, u, playClip)
		}(topic)
	}

	<-ctx.Done()      // wait for Ctrl-C
	log.Println("shutting down …")
	speaker.Close()   // close audio device immediately
	wg.Wait()         // all goroutines have exited
}
