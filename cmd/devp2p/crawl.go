// Copyright 2019 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"time"

	"github.com/youngqqcn/arbitrum/log"
	"github.com/youngqqcn/arbitrum/p2p/enode"
)

type crawler struct {
	input     nodeSet
	output    nodeSet
	disc      resolver
	iters     []enode.Iterator
	inputIter enode.Iterator
	ch        chan *enode.Node
	closed    chan struct{}

	// settings
	revalidateInterval time.Duration
}

const (
	nodeRemoved = iota
	nodeSkipRecent
	nodeSkipIncompat
	nodeAdded
	nodeUpdated
)

type resolver interface {
	RequestENR(*enode.Node) (*enode.Node, error)
}

func newCrawler(input nodeSet, disc resolver, iters ...enode.Iterator) *crawler {
	c := &crawler{
		input:     input,
		output:    make(nodeSet, len(input)),
		disc:      disc,
		iters:     iters,
		inputIter: enode.IterNodes(input.nodes()),
		ch:        make(chan *enode.Node),
		closed:    make(chan struct{}),
	}
	c.iters = append(c.iters, c.inputIter)
	// Copy input to output initially. Any nodes that fail validation
	// will be dropped from output during the run.
	for id, n := range input {
		c.output[id] = n
	}
	return c
}

func (c *crawler) run(timeout time.Duration) nodeSet {
	var (
		timeoutTimer = time.NewTimer(timeout)
		timeoutCh    <-chan time.Time
		statusTicker = time.NewTicker(time.Second * 8)
		doneCh       = make(chan enode.Iterator, len(c.iters))
		liveIters    = len(c.iters)
	)
	defer timeoutTimer.Stop()
	defer statusTicker.Stop()
	for _, it := range c.iters {
		go c.runIterator(doneCh, it)
	}

	var (
		added   int
		updated int
		skipped int
		recent  int
		removed int
	)
loop:
	for {
		select {
		case n := <-c.ch:
			switch c.updateNode(n) {
			case nodeSkipIncompat:
				skipped++
			case nodeSkipRecent:
				recent++
			case nodeRemoved:
				removed++
			case nodeAdded:
				added++
			default:
				updated++
			}
		case it := <-doneCh:
			if it == c.inputIter {
				// Enable timeout when we're done revalidating the input nodes.
				log.Info("Revalidation of input set is done", "len", len(c.input))
				if timeout > 0 {
					timeoutCh = timeoutTimer.C
				}
			}
			if liveIters--; liveIters == 0 {
				break loop
			}
		case <-timeoutCh:
			break loop
		case <-statusTicker.C:
			log.Info("Crawling in progress",
				"added", added, "updated", updated, "removed", removed,
				"ignored(recent)", recent, "ignored(incompatible)", skipped)
		}
	}

	close(c.closed)
	for _, it := range c.iters {
		it.Close()
	}
	for ; liveIters > 0; liveIters-- {
		<-doneCh
	}
	return c.output
}

func (c *crawler) runIterator(done chan<- enode.Iterator, it enode.Iterator) {
	defer func() { done <- it }()
	for it.Next() {
		select {
		case c.ch <- it.Node():
		case <-c.closed:
			return
		}
	}
}

// updateNode updates the info about the given node, and returns a status
// about what changed
func (c *crawler) updateNode(n *enode.Node) int {
	node, ok := c.output[n.ID()]

	// Skip validation of recently-seen nodes.
	if ok && time.Since(node.LastCheck) < c.revalidateInterval {
		return nodeSkipRecent
	}

	// Request the node record.
	nn, err := c.disc.RequestENR(n)
	node.LastCheck = truncNow()
	status := nodeUpdated
	if err != nil {
		if node.Score == 0 {
			// Node doesn't implement EIP-868.
			log.Debug("Skipping node", "id", n.ID())
			return nodeSkipIncompat
		}
		node.Score /= 2
	} else {
		node.N = nn
		node.Seq = nn.Seq()
		node.Score++
		if node.FirstResponse.IsZero() {
			node.FirstResponse = node.LastCheck
			status = nodeAdded
		}
		node.LastResponse = node.LastCheck
	}

	// Store/update node in output set.
	if node.Score <= 0 {
		log.Debug("Removing node", "id", n.ID())
		delete(c.output, n.ID())
		return nodeRemoved
	}
	log.Debug("Updating node", "id", n.ID(), "seq", n.Seq(), "score", node.Score)
	c.output[n.ID()] = node
	return status
}

func truncNow() time.Time {
	return time.Now().UTC().Truncate(1 * time.Second)
}
