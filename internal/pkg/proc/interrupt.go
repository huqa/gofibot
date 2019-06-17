// Package proc provides utility functions for dealing with processes
package proc

import (
    "os"
    "os/signal"
)

func WaitForInterrupt() {
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt)
    <-stop
}
