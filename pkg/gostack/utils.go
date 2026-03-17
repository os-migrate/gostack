/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2025 Red Hat, Inc.
 *
 */
package gostack

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"
)

// RandomID generates a UUID v4-ish formatted ID with the given prefix.
func RandomID(prefix string) string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf(
		"%s-%x-%x-%x-%x-%x",
		prefix,
		b[0:4],
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:16],
	)
}

// BuildAddresses builds network addresses from server networks.
func BuildAddresses(networks []map[string]string) map[string][]map[string]interface{} {
	addresses := map[string][]map[string]interface{}{}
	for i, net := range networks {
		netName := "private"
		if netID, ok := net["uuid"]; ok {
			netName = netID
		}
		addresses[netName] = append(addresses[netName], map[string]interface{}{
			"addr":                    fmt.Sprintf("192.168.0.%d", 10+i),
			"version":                 4,
			"OS-EXT-IPS:type":         "fixed",
			"OS-EXT-IPS-MAC:mac_addr": "fa:16:3e:00:00:01",
		})
	}
	return addresses
}

// Parts splits a path into path segments.
func Parts(p string) []string {
	return strings.Split(strings.Trim(p, "/"), "/")
}

// ProjectIDFromToken extracts project ID from request headers or token.
func ProjectIDFromToken(r *http.Request) string {
	token := r.Header.Get("X-Auth-Token")
	if token == "" {
		return "demo"
	}
	tokenMu.RLock()
	defer tokenMu.RUnlock()
	if projectID, ok := tokenProjects[token]; ok {
		return projectID
	}
	if strings.HasPrefix(token, "fake-token-") {
		proj := strings.TrimPrefix(token, "fake-token-")
		if p, ok := projects[proj]; ok {
			return p.ID
		}
		for _, p := range projects {
			if p.Name == proj {
				return p.ID
			}
		}
	}
	if projID := r.Header.Get("X-Project-Id"); projID != "" {
		if p, ok := projects[projID]; ok {
			return p.ID
		}
	}
	if projID := r.Header.Get("X-Auth-Project-Id"); projID != "" {
		if p, ok := projects[projID]; ok {
			return p.ID
		}
	}
	if projName := r.Header.Get("X-Project-Name"); projName != "" {
		for _, p := range projects {
			if p.Name == projName {
				return p.ID
			}
		}
	}
	if projName := r.Header.Get("X-Auth-Project-Name"); projName != "" {
		for _, p := range projects {
			if p.Name == projName {
				return p.ID
			}
		}
	}
	return "demo"
}
