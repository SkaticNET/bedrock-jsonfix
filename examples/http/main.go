package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/SkaticNET/bedrock-jsonfix/bedrockjsonfix"
)

func main() {
	http.HandleFunc("/fix", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		opt := bedrockjsonfix.DefaultOptions()
		opt.MaxInputBytes = 2 << 20
		res, err := bedrockjsonfix.FixReader(ctx, r.Body, opt)
		if err != nil {
			if errors.Is(err, bedrockjsonfix.ErrInputTooLarge) {
				http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
				return
			}
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(res.Output); err != nil {
			log.Printf("write response: %v", err)
			return
		}
	})
	fmt.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
