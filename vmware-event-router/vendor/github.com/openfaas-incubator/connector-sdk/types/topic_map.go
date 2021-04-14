// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"sync"
)

func NewTopicMap() TopicMap {
	lookup := make(map[string][]string)
	return TopicMap{
		lookup: &lookup,
		lock:   sync.RWMutex{},
	}
}

type TopicMap struct {
	lookup *map[string][]string
	lock   sync.RWMutex
}

func (t *TopicMap) Match(topicName string) []string {
	t.lock.RLock()
	defer t.lock.RUnlock()

	var values []string

	for key, val := range *t.lookup {
		if key == topicName {
			values = val
			break
		}
	}

	return values
}

func (t *TopicMap) Sync(updated *map[string][]string) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.lookup = updated
}

func (t *TopicMap) Topics() []string {
	t.lock.RLock()
	defer t.lock.RUnlock()

	topics := make([]string, 0, len(*t.lookup))
	for topic := range *t.lookup {
		topics = append(topics, topic)
	}

	return topics
}
