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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// RegisterCinderHandlers registers Cinder (volume) API handlers on mux.
func RegisterCinderHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v3/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.String())
		p := Parts(r.URL.Path)
		// Match /v3/{project}/volumes/detail first
		if strings.HasSuffix(r.URL.Path, "/detail") {
			list := []interface{}{}
			for _, v := range volumes {
				list = append(list, v)
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(map[string]interface{}{
				"volumes": list,
			})
			if err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		// GET /v3/{project}/volumes/{id}
		if len(p) == 4 && p[2] == "volumes" && r.Method == http.MethodGet {
			volID := p[3]

			mu.Lock()
			vol, ok := volumes[volID]
			mu.Unlock()

			if !ok {
				http.NotFound(w, r)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(map[string]interface{}{"volume": vol})
			if err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		// POST /v3/{project}/volumes/{id}/action
		if len(p) == 5 && p[2] == "volumes" && p[4] == "action" && r.Method == http.MethodPost {
			volID := p[3]

			mu.Lock()
			vol, ok := volumes[volID]
			mu.Unlock()

			if !ok {
				http.NotFound(w, r)
				return
			}

			// Parse the action body
			var actionBody map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&actionBody); err != nil {
				http.Error(w, "invalid JSON body", http.StatusBadRequest)
				return
			}

			// Handle os-set_bootable action
			if bootableData, ok := actionBody["os-set_bootable"]; ok {
				if bootableMap, ok := bootableData.(map[string]interface{}); ok {
					if bootable, ok := bootableMap["bootable"].(bool); ok {
						mu.Lock()
						if vol.Metadata == nil {
							vol.Metadata = make(map[string]string)
						}
						vol.Metadata["bootable"] = fmt.Sprintf("%t", bootable)
						mu.Unlock()
						log.Printf("Set volume %s bootable=%t", volID, bootable)
					}
				}
				w.WriteHeader(http.StatusOK)
				return
			}

			// Handle os-set_image_metadata action (for UEFI firmware, etc.)
			if imageMetadata, ok := actionBody["os-set_image_metadata"]; ok {
				if metadataMap, ok := imageMetadata.(map[string]interface{}); ok {
					if metadata, ok := metadataMap["metadata"].(map[string]interface{}); ok {
						mu.Lock()
						if vol.Metadata == nil {
							vol.Metadata = make(map[string]string)
						}
						for k, v := range metadata {
							if strVal, ok := v.(string); ok {
								vol.Metadata[k] = strVal
							}
						}
						mu.Unlock()
						log.Printf("Set volume %s image metadata: %v", volID, metadata)
					}
				}
				w.WriteHeader(http.StatusOK)
				return
			}

			// Unknown action
			http.Error(w, "Unsupported action", http.StatusBadRequest)
			return
		}
		// DELETE /v3/{project}/volumes/{id}
		if len(p) == 4 && p[2] == "volumes" && r.Method == http.MethodDelete {
			volID := p[3]

			mu.Lock()
			vol, ok := volumes[volID]
			mu.Unlock()

			if !ok {
				http.NotFound(w, r)
				return
			}

			if len(vol.Attachments) > 0 {
				http.Error(w, "Volume is attached to a server", http.StatusBadRequest)
				return
			}

			if vol.Status != "available" && vol.Status != "error" {
				http.Error(w, "Volume must be available or error to delete", http.StatusBadRequest)
				return
			}

			mu.Lock()
			delete(volumes, volID)
			mu.Unlock()

			log.Printf("Deleted volume %s", volID)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if len(p) == 3 && p[2] == "volumes" {
			if r.Method == http.MethodGet {
				q := r.URL.Query()
				nameFilter := q.Get("name")
				statusFilter := q.Get("status")

				metadataFilters := make(map[string]string)
				for key, values := range q {
					if strings.HasPrefix(key, "metadata[") && strings.HasSuffix(key, "]") {
						metaKey := key[9 : len(key)-1]
						if len(values) > 0 {
							metadataFilters[metaKey] = values[0]
						}
					}
				}

				mu.Lock()
				list := []interface{}{}
				for _, v := range volumes {
					if nameFilter != "" && v.Name != nameFilter {
						continue
					}
					if statusFilter != "" && v.Status != statusFilter {
						continue
					}
					matchesMeta := true
					for metaKey, metaValue := range metadataFilters {
						if v.Metadata[metaKey] != metaValue {
							matchesMeta = false
							break
						}
					}
					if !matchesMeta {
						continue
					}
					list = append(list, v)
				}
				mu.Unlock()

				err := json.NewEncoder(w).Encode(map[string]interface{}{"volumes": list})
				if err != nil {
					log.Printf("Error encoding JSON response: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
				return
			}
			if r.Method == http.MethodPost {
				var req struct {
					Volume struct {
						Name     string            `json:"name"`
						Size     int               `json:"size"`
						Metadata map[string]string `json:"metadata"`
					} `json:"volume"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, "invalid JSON body", http.StatusBadRequest)
					return
				}

				id := fmt.Sprintf("%d", volumeID)
				volumeID++

				v := &Volume{
					ID:          id,
					Name:        req.Volume.Name,
					Size:        req.Volume.Size,
					Status:      "creating",
					Attachments: []map[string]interface{}{},
					Metadata:    req.Volume.Metadata,
				}

				if v.Metadata == nil {
					v.Metadata = map[string]string{}
				}

				mu.Lock()
				volumes[id] = v
				mu.Unlock()

				go func(volID string) {
					time.Sleep(2 * time.Second)
					mu.Lock()
					defer mu.Unlock()
					if vol, ok := volumes[volID]; ok {
						vol.Status = "available"
						log.Printf("Volume %s transitioned to available", volID)
					}
				}(id)

				w.WriteHeader(http.StatusAccepted)
				err := json.NewEncoder(w).Encode(map[string]interface{}{"volume": v})
				if err != nil {
					log.Printf("Error encoding JSON response: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
				return
			}
			if r.Method == http.MethodDelete {
				http.Error(w, "DELETE volumes list not supported", http.StatusMethodNotAllowed)
				return
			}
			http.Error(w, "Method not Implemented", http.StatusNotImplemented)
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
}
