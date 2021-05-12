package main

import (
	"bytes"
	"context"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	firestore "cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func startFirebaseEmulator(t *testing.T, ctx context.Context) *firestore.Client {
	t.Helper()
	port := "8010"
	addr := "localhost:" + port
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(func() {
		t.Log("shutting down firestore operator")
		cancel()
	})

	cmd := exec.CommandContext(ctx, "gcloud", "beta", "emulators", "firestore", "start", "--host-port="+addr)
	var out bytes.Buffer
	cmd.Stdin, cmd.Stdout = &out, &out
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start firestore emulator: %v -- out:%s", err, out.String())
	}
	dialCtx, clean := context.WithTimeout(ctx, time.Second*10)
	defer clean()
	var connected bool
	for !connected {
		select {
		case <-dialCtx.Done():
			t.Fatalf("emulator did not come up timely: %v -- output: %s", dialCtx.Err(), out.String())
		default:
			c, err := (&net.Dialer{Timeout: time.Millisecond * 200}).DialContext(ctx, "tcp", addr)
			if err == nil {
				c.Close()
				t.Log("firestore emulator started")
				connected = true
				break
			}
			time.Sleep(time.Millisecond * 200) //before retrying
		}
	}
	os.Setenv("FIRESTORE_EMULATOR_HOST", addr)
	cl, err := firestore.NewClient(ctx, firestore.DetectProjectID)
	if err != nil {
		t.Fatal(err)
	}
	os.Unsetenv("FIRESTORE_EMULATOR_HOST")
	clearDB(t, ctx, cl)
	return cl
}

func clearDB(t *testing.T, ctx context.Context, cl *firestore.Client) {
	t.Helper()
	cs, err := cl.Collections(ctx).GetAll()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cs {
		it := c.Limit(1000).Documents(ctx)
		for {
			deleted := 0
			batch := cl.Batch()
			for {
				doc, err := it.Next()
				if err == iterator.Done {
					break
				} else if err != nil {
					t.Fatal(err)
				}
				batch.Delete(doc.Ref)
				deleted++
			}
			if deleted == 0 {
				return
			}
			if _, err := batch.Commit(ctx); err != nil {
				t.Fatal(err)
			}
		}
	}
}
