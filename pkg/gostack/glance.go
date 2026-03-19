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
)

// RegisterGlanceHandlers registers Glance (image) API handlers on mux.
func RegisterGlanceHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v2/images", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			projectID := ProjectIDFromToken(r)
			nameFilter := r.URL.Query().Get("name")
			ownerFilter := r.URL.Query().Get("owner")
			projectImages := []map[string]interface{}{}
			publicImages := []map[string]interface{}{}
			mu.Lock()
			for _, img := range images {
				visibility, _ := img["visibility"].(string)
				owner, _ := img["owner"].(string)
				if nameFilter != "" {
					if n, ok := img["name"].(string); !ok || n != nameFilter {
						continue
					}
				}
				if ownerFilter != "" && owner != ownerFilter {
					continue
				}
				if owner == projectID {
					projectImages = append(projectImages, img)
				} else if visibility == "public" {
					publicImages = append(publicImages, img)
				}
			}
			mu.Unlock()
			visible := []map[string]interface{}{}
			seenNames := make(map[string]bool)
			for _, img := range projectImages {
				name, _ := img["name"].(string)
				seenNames[name] = true
				visible = append(visible, img)
			}
			for _, img := range publicImages {
				name, _ := img["name"].(string)
				if !seenNames[name] {
					visible = append(visible, img)
				}
			}
			err := json.NewEncoder(w).Encode(map[string]interface{}{"images": visible})
			if err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		if r.Method == http.MethodPost {
			var img map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&img); err != nil {
				http.Error(w, "invalid JSON body", http.StatusBadRequest)
				return
			}
			img["id"] = fmt.Sprintf("img-%d", len(images)+1)
			img["status"] = "active"
			if _, ok := img["owner"]; !ok {
				img["owner"] = ProjectIDFromToken(r)
			}
			if _, ok := img["visibility"]; !ok {
				img["visibility"] = "private"
			}
			mu.Lock()
			images = append(images, img)
			mu.Unlock()
			w.WriteHeader(http.StatusCreated)
			err := json.NewEncoder(w).Encode(img)
			if err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/v2/images/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := r.URL.Path[len("/v2/images/"):]
		projectID := ProjectIDFromToken(r)

		mu.Lock()
		defer mu.Unlock()
		for _, img := range images {
			if imgID, ok := img["id"].(string); ok && imgID == id {
				visibility, _ := img["visibility"].(string)
				owner, _ := img["owner"].(string)
				if visibility == "public" || owner == projectID {
					err := json.NewEncoder(w).Encode(img)
					if err != nil {
						log.Printf("Error encoding JSON response: %v", err)
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					}
					return
				}
				http.NotFound(w, r)
				return
			}
		}
		http.NotFound(w, r)
	})
}
