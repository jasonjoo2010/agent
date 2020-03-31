package engine

import (
	"bufio"
	"context"
	"io"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/docker/docker/pkg/stdcopy"

	dockertypes "github.com/docker/docker/api/types"
	coreutils "github.com/projecteru2/core/utils"
	log "github.com/sirupsen/logrus"

	"github.com/projecteru2/agent/common"
	"github.com/projecteru2/agent/types"
	"github.com/projecteru2/agent/watcher"
)

func (e *Engine) pumpToWriter(typ string, source io.Reader, container *types.Container) {
	buf := bufio.NewReader(source)
	for {
		data, err := buf.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Errorf("[attach] attach pumpToWriter %s %s %s %s", container.Name, coreutils.ShortID(container.ID), typ, err)
			}
			return
		}
		data = strings.TrimSuffix(data, "\n")
		data = strings.TrimSuffix(data, "\r")
		l := &types.Log{
			ID:         container.ID,
			Name:       container.Name,
			Type:       typ,
			EntryPoint: container.EntryPoint,
			Ident:      container.Ident,
			Data:       data,
			Datetime:   time.Now().Format(common.DateTimeFormat),
			//TODO
			//Extra
		}
		watcher.GetInstance().LogC <- l
		if err := e.writer.Write(l); err != nil && !(container.EntryPoint == "agent" && e.dockerized) {
			// log.Errorf("[attach] %s container %s_%s write failed %v", container.Name, container.EntryPoint, coreutils.ShortID(container.ID), err)
			// log.Errorf("[attach] %s", data)
			// XXX: In favor of writer.{Dropped(), Sent(), Failed()} it's suggested
			// that use other approach to do this. An example is given here.
			if e.writer.Dropped() == 1000 || e.writer.Failed() == 1000 {
				log.Errorf("[attach] Sending to forwarders fail: %d sent, %d dropped, %d failed",
					e.writer.Sent(), e.writer.Dropped, e.writer.Failed())
			}
		}
	}
}

func (e *Engine) drainLogs(ctx context.Context, cancel context.CancelFunc,
	container *types.Container,
	outWriter, errWriter *io.PipeWriter) {
	options := dockertypes.ContainerAttachOptions{
		Stream: true,
		Stdin:  false,
		Stdout: true,
		Stderr: true,
	}
	resp, err := e.docker.ContainerAttach(ctx, container.ID, options)
	if err != nil && err != httputil.ErrPersistEOF {
		log.Errorf("[attach] attach %s container %s failed %s", container.Name, coreutils.ShortID(container.ID), err)
		return
	}
	defer resp.Close()
	defer outWriter.Close()
	defer errWriter.Close()
	defer cancel()
	_, err = stdcopy.StdCopy(outWriter, errWriter, resp.Reader)
	if err != nil {
		log.Errorf("[attach] attach get stream failed %s", err)
	}
	log.Infof("[attach] attach %s container %s finished", container.Name, coreutils.ShortID(container.ID))
}

func (e *Engine) attach(container *types.Container) {
	outr, outw := io.Pipe()
	errr, errw := io.Pipe()
	ctx := context.Background()
	cancelCtx, cancel := context.WithCancel(ctx)
	go e.drainLogs(ctx, cancel, container, outw, errw)
	log.Infof("[attach] attach %s container %s success", container.Name, coreutils.ShortID(container.ID))
	// attach metrics
	go e.stat(cancelCtx, container)

	go e.pumpToWriter("stdout", outr, container)
	go e.pumpToWriter("stderr", errr, container)
}
