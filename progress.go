package uiprogress

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/gosuri/uilive"
)

// Out is the default writer to render progress bars to
var Out = os.Stdout

// RefreshInterval in the default time duration to wait for refreshing the output
var RefreshInterval = time.Millisecond * 10

// defaultProgress is the default progress
var defaultProgress = New()

// Progress represents the container that renders progress bars
type Progress struct {
	// Out is the writer to render progress bars to
	Out io.Writer

	// Width is the width of the progress bars
	Width int

	// Bars is the collection of progress bars
	Bars []*Bar

	// RefreshInterval in the time duration to wait for refreshing the output
	RefreshInterval time.Duration

	lw     *uilive.Writer
	ticker *time.Ticker
	tdone  chan bool
	mtx    *sync.RWMutex
}

// New returns a new progress bar with defaults
func New() *Progress {
	lw := uilive.New()
	lw.Out = Out

	return &Progress{
		Width:           Width,
		Out:             Out,
		Bars:            make([]*Bar, 0),
		RefreshInterval: RefreshInterval,

		tdone: make(chan bool),
		lw:    lw,
		mtx:   &sync.RWMutex{},
	}
}

// AddBar creates a new progress bar and adds it to the default progress container
func AddBar(total int) *Bar {
	return defaultProgress.AddBar(total)
}

// RemoveBar removes the progress bar from the container and returns true, else it returns false
func RemoveBar(bar *Bar) bool {
	return defaultProgress.RemoveBar(bar)
}

// Start starts the rendering the progress of progress bars using the DefaultProgress. It listens for updates using `bar.Set(n)` and new bars when added using `AddBar`
func Start() {
	defaultProgress.Start()
}

// Stop stops listening
func Stop() {
	defaultProgress.Stop()
}

// Listen listens for updates and renders the progress bars
func Listen() {
	defaultProgress.Listen()
}

func Print() {
	defaultProgress.Print()
}

// SetOut sets the `io.Writer` to use for output
func (p *Progress) SetOut(o io.Writer) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	p.Out = o
	p.lw.Out = o
}

// SetRefreshInterval sets the refresh interval
func (p *Progress) SetRefreshInterval(interval time.Duration) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.RefreshInterval = interval
}

// AddBar creates a new progress bar and adds to the container
func (p *Progress) AddBar(total int) *Bar {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	bar := NewBar(total)
	bar.Width = p.Width
	p.Bars = append(p.Bars, bar)
	return bar
}

// RemoveBar removes the progress bar from the container and returns true, else it returns false
func (p *Progress) RemoveBar(bar *Bar) bool {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if i := p.indexOf(bar); i != -1 {
		p.Bars = append(p.Bars[:i], p.Bars[i+1:]...)
		return true
	}
	return false
}

func (p *Progress) indexOf(bar *Bar) int {
	for i, b := range p.Bars {
		if bar == b {
			return i
		}
	}
	return -1
}

// Listen listens for updates and renders the progress bars
func (p *Progress) Listen() {
	for {
		p.mtx.Lock()
		interval := p.RefreshInterval
		p.mtx.Unlock()

		select {
		case <-time.After(interval):
			p.Print()
		case <-p.tdone:
			p.Print()
			close(p.tdone)
			return
		}
	}
}

func (p *Progress) Print() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	for _, bar := range p.Bars {
		fmt.Fprintln(p.lw, bar.String())
	}
	p.lw.Flush()
}

// Start starts the rendering the progress of progress bars. It listens for updates using `bar.Set(n)` and new bars when added using `AddBar`
func (p *Progress) Start() {
	go p.Listen()
}

// Stop stops listening
func (p *Progress) Stop() {
	p.tdone <- true
	<-p.tdone
}

// Bypass returns a writer which allows non-buffered data to be written to the underlying output
func (p *Progress) Bypass() io.Writer {
	return p.lw.Bypass()
}
