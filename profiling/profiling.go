/*
Package profiling provides tools for recording frame timings for later analysis.

The simplest usage is to construct a profilier at the start of your event loop
function and to defer stopping it until the window is closed. Something like:

    func loop(w *app.Window) error {
        // log to a CSV file with a randomly-chosen name. The file's path will be
        // logged to stderr.
        recorder, err := NewRecorder(nil)
        if err != nil {
            // handle
        }
        defer recorder.Stop()

        var ops op.Ops

        for event := range window.Events() {
            switch event := event.(type) {
                case system.DestroyEvent:
                    // returning will execute the deferred call to Stop(), which
                    // flushes the CSV file.
                    return event.Err
                case system.FrameEvent:
                    gtx := layout.NewContext(&ops, event)

                    // record the last frame's timing info and prepare the next one
                    recorder.profile(gtx)

                    // lay out your UI here
            }
        }
    }

The actual disk I/O is performed by a different goroutine and will not block the UI unless it gets more than 60 frames behind.
*/
package profiling

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"gioui.org/io/profile"
	"gioui.org/layout"
)

// Timings holds frame timing information
type Timings struct {
	// When is the timestamp of the frame that this struct describes
	When          time.Time
	Total         time.Duration
	FrameDuration time.Duration
	GPUTime       time.Duration
	ZT            time.Duration
	ST            time.Duration
	CovT          time.Duration
}

func (t Timings) CSVRow() []string {
	out := []string{}
	toString := func(d time.Duration) string {
		return strconv.Itoa(int(d.Milliseconds()))
	}
	out = append(out, strconv.Itoa(int(t.When.UnixNano()/1e6)))
	out = append(out, toString(t.Total))
	out = append(out, toString(t.FrameDuration))
	out = append(out, toString(t.GPUTime))
	out = append(out, toString(t.ZT))
	out = append(out, toString(t.ST))
	out = append(out, toString(t.CovT))
	return out
}

func decode(e timedEvent) Timings {
	out := Timings{}
	var tot, fd, gpu, zt, st, covt string
	fmt.Sscanf(e.Timings, "tot: %s draw: %s gpu: %s zt: %s st: %s cov: %s", &tot, &fd, &gpu, &zt, &st, &covt)
	out.When = e.when
	out.Total, _ = time.ParseDuration(tot)
	out.FrameDuration, _ = time.ParseDuration(fd)
	out.GPUTime, _ = time.ParseDuration(gpu)
	out.ZT, _ = time.ParseDuration(zt)
	out.ST, _ = time.ParseDuration(st)
	out.CovT, _ = time.ParseDuration(covt)
	return out
}

func header() []string {
	return []string{"when(unix ms)", "tot(ms)", "draw(ms)", "gpu(ms)", "zt(ms)", "st(ms)", "cov(ms)"}
}

type timedEvent struct {
	profile.Event
	when time.Time
}

// CSVTimingRecorder captures frame timing information into a CSV file
type CSVTimingRecorder struct {
	nextEventTime time.Time
	file          *os.File
	csvWriter     *csv.Writer
	listener      chan timedEvent
	errChan       chan error
}

// NewRecorder creates a CSVTimingRecorder that will record to a CSV file
// with the provided name. If the name is nil, a temporary file will be used.
func NewRecorder(filename *string) (*CSVTimingRecorder, error) {
	var (
		err  error
		file *os.File
	)
	if filename == nil {
		file, err = ioutil.TempFile("", "profile-*.csv")
	} else {
		file, err = os.Create(*filename)
	}
	if err != nil {
		return nil, fmt.Errorf("failed opening csv file: %w", err)
	}
	recorder := &CSVTimingRecorder{}
	recorder.file = file
	recorder.csvWriter = csv.NewWriter(recorder.file)
	recorder.listener = make(chan timedEvent, 60)
	recorder.errChan = make(chan error)

	go recorder.consume()
	return recorder, nil
}

func (c *CSVTimingRecorder) consume() {
	defer close(c.errChan)
	log.Printf("Logging csv profiling to %v", c.file.Name())
	c.csvWriter.Write(header())
	for e := range c.listener {
		timing := decode(e)
		err := c.csvWriter.Write(timing.CSVRow())
		if err != nil {
			c.errChan <- err
		}
	}
	c.csvWriter.Flush()
	c.errChan <- c.csvWriter.Error()
}

// Stop shuts down the recording process and flushes all data to the
// CSV file.
func (c *CSVTimingRecorder) Stop() error {
	close(c.listener)
	err := <-c.errChan
	if err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}
	err = c.file.Close()
	if err != nil {
		return fmt.Errorf("failed to close csv: %w", err)
	}
	log.Printf("CSV profiling data written to %s", c.file.Name())
	return nil
}

// Profile records profiling data from the last frame and prepares the
// capture of the next frame. Calling this method every frame is sufficient
// to profile all frames. It will simply return if c is nil.
func (c *CSVTimingRecorder) Profile(gtx layout.Context) {
	if c == nil {
		return
	}
	var lastEventTime time.Time
	lastEventTime, c.nextEventTime = c.nextEventTime, gtx.Now
	profile.Op{Tag: c}.Add(gtx.Ops)
	for _, e := range gtx.Events(c) {
		switch e := e.(type) {
		case profile.Event:
			c.Write(lastEventTime, e)
		}
	}
}

// Write is a lower-level way to capture a single profile event. It should
// be used instead of the Profile method if more granular profiling control
// is desired. It will simply return if c is nil.
func (c *CSVTimingRecorder) Write(when time.Time, e profile.Event) error {
	if c == nil {
		return nil
	}
	var err error
	select {
	case err = <-c.errChan:
	default:
	}
	select {
	case c.listener <- timedEvent{Event: e, when: when}:
	default:
		err = fmt.Errorf("recorder already stopped")
	}
	return err
}
