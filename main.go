// Copyright 2025 Laurynas Četyrkinas <laurynas@digilol.net>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/digilolnet/digilol-cert-pushpuller/internal/config"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		fmt.Println("Usage: digilol-cert-pushpuller <push|pull> --config /path/to/config.toml")
		os.Exit(1)
	}

	command := os.Args[1]

	// Parse common flags
	var configPath string
	fs := flag.NewFlagSet(command, flag.ExitOnError)
	fs.StringVar(&configPath, "config", "", "Path to config file")
	fs.Parse(os.Args[2:])

	if configPath == "" {
		log.Fatal("--config is required")
	}

	switch command {
	case "push":
		cfg, err := config.LoadPush(configPath)
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}

		if cfg.Daemon.Enabled {
			runPushDaemon(cfg)
		} else {
			if err := push(cfg); err != nil {
				log.Fatalf("push failed: %v", err)
			}
		}

	case "pull":
		cfg, err := config.LoadPull(configPath)
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}

		if cfg.Daemon.Enabled {
			runPullDaemon(cfg)
		} else {
			if err := pull(cfg); err != nil {
				log.Fatalf("pull failed: %v", err)
			}
		}

	default:
		log.Fatalf("unknown command: %s", command)
	}
}

func runDaemon(name string, intervalSecs, jitterSecs int, fn func() error) {
	if jitterSecs > 0 {
		log.Printf("starting %s daemon (interval: %ds, jitter: %ds)", name, intervalSecs, jitterSecs)
	} else {
		log.Printf("starting %s daemon (interval: %ds)", name, intervalSecs)
	}

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Seed random for jitter
	if jitterSecs > 0 {
		rand.Seed(time.Now().UnixNano())
	}

	// Run immediately on startup
	if err := fn(); err != nil {
		log.Printf("%s failed: %v", name, err)
	}

	ticker := time.NewTicker(time.Duration(intervalSecs) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Apply jitter if configured
			if jitterSecs > 0 {
				jitter := time.Duration(rand.Intn(jitterSecs)) * time.Second
				time.Sleep(jitter)
			}

			if err := fn(); err != nil {
				log.Printf("%s failed: %v", name, err)
			}

		case sig := <-sigChan:
			log.Printf("received signal %v, shutting down", sig)
			return
		}
	}
}

func runPushDaemon(cfg *config.PushConfig) {
	runDaemon("push", cfg.Daemon.IntervalSecs, cfg.Daemon.JitterSecs, func() error {
		return push(cfg)
	})
}

func runPullDaemon(cfg *config.PullConfig) {
	runDaemon("pull", cfg.Daemon.IntervalSecs, cfg.Daemon.JitterSecs, func() error {
		return pull(cfg)
	})
}
