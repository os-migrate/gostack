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
	"path"
)

// RegisterKeystoneHandlers registers Keystone API handlers on mux.
func RegisterKeystoneHandlers(mux *http.ServeMux, roleAssignments map[string]map[string][]string) {
	mux.HandleFunc("/v3", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.String())
		log.Printf("%s %s", r.URL.Path, r.URL.RawQuery)
		err := json.NewEncoder(w).Encode(map[string]interface{}{
			"version": map[string]interface{}{
				"id":     "v3.14",
				"status": "stable",
				"links": []map[string]string{
					{"rel": "self", "href": baseURL + "/v3/"},
				},
			},
		})
		if err != nil {
			log.Printf("Error encoding JSON response: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/v3/auth/tokens", func(w http.ResponseWriter, r *http.Request) {
		projectID := "source"
		projectName := "source"
		if r.Body != nil {
			var authReq struct {
				Auth struct {
					Scope struct {
						Project struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"project"`
					} `json:"scope"`
				} `json:"auth"`
			}
			if err := json.NewDecoder(r.Body).Decode(&authReq); err == nil {
				if authReq.Auth.Scope.Project.ID != "" {
					projectID = authReq.Auth.Scope.Project.ID
				}
				if authReq.Auth.Scope.Project.Name != "" {
					projectName = authReq.Auth.Scope.Project.Name
				}
			}
		}
		if p, ok := projects[projectID]; ok {
			projectID = p.ID
			projectName = p.Name
		} else if p, ok := projects[projectName]; ok {
			projectID = p.ID
			projectName = p.Name
		} else {
			for _, p := range projects {
				if p.Name == projectName {
					projectID = p.ID
					projectName = p.Name
					break
				}
			}
		}
		token := "fake-token-" + projectID
		w.Header().Set("X-Subject-Token", token)
		tokenMu.Lock()
		tokenProjects[token] = projectID
		tokenMu.Unlock()
		w.WriteHeader(http.StatusCreated)
		catalog := []map[string]interface{}{
			{
				"type": "identity",
				"name": "keystone",
				"endpoints": []map[string]interface{}{
					{
						"id": "keystone-public", "interface": "public",
						"region": "RegionOne", "region_id": "RegionOne",
						"url": baseURL + "/v3/",
					},
				},
			},
			{
				"type": "compute",
				"name": "nova",
				"endpoints": []map[string]interface{}{
					{
						"id": "nova-public", "interface": "public",
						"region": "RegionOne", "region_id": "RegionOne",
						"url": baseURL + "/v2.1/" + projectID,
					},
				},
			},
			{
				"type": "network",
				"name": "neutron",
				"endpoints": []map[string]interface{}{
					{
						"id": "neutron-public", "interface": "public",
						"region": "RegionOne", "region_id": "RegionOne",
						"url": baseURL + "/v2.0",
					},
				},
			},
			{
				"type": "image",
				"name": "glance",
				"endpoints": []map[string]interface{}{
					{
						"id": "glance-public", "interface": "public",
						"region": "RegionOne", "region_id": "RegionOne",
						"url": baseURL + "/v2",
					},
				},
			},
			{
				"type": "volumev3",
				"name": "cinder",
				"endpoints": []map[string]interface{}{
					{
						"id": "cinder-public", "interface": "public",
						"region": "RegionOne", "region_id": "RegionOne",
						"url": baseURL + "/v3/" + projectID,
					},
				},
			},
		}
		err := json.NewEncoder(w).Encode(map[string]interface{}{
			"token": map[string]interface{}{
				"expires_at": "2099-01-01T00:00:00.000000Z",
				"issued_at":  "2026-01-01T00:00:00.000000Z",
				"project": map[string]string{"id": projectID, "name": projectName},
				"user": map[string]string{"id": "demo-user", "name": "demo"},
				"catalog": catalog,
			},
		})
		if err != nil {
			log.Printf("Error encoding JSON response: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/v3/domains", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		log.Printf("%s %s", r.Method, r.URL.String())
		if r.Method == http.MethodGet {
			list := []interface{}{}
			for _, d := range domains {
				list = append(list, d)
			}
			err := json.NewEncoder(w).Encode(map[string]interface{}{"domains": list})
			if err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		if r.Method == http.MethodPost {
			id := fmt.Sprintf("domain-%d", seq)
			seq++
			domains[id] = &Domain{ID: id, Name: id}
			w.WriteHeader(http.StatusCreated)
			err := json.NewEncoder(w).Encode(map[string]interface{}{"domain": domains[id]})
			if err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/v3/domains/", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		log.Printf("%s %s", r.Method, r.URL.String())
		if r.Method == http.MethodGet {
			list := []interface{}{}
			for _, d := range domains {
				list = append(list, d)
			}
			err := json.NewEncoder(w).Encode(map[string]interface{}{"domains": list})
			if err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		if r.Method == http.MethodPost {
			id := fmt.Sprintf("domain-%d", seq)
			seq++
			domains[id] = &Domain{ID: id, Name: id}
			w.WriteHeader(http.StatusCreated)
			err := json.NewEncoder(w).Encode(map[string]interface{}{"domain": domains[id]})
			if err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
	})

	mux.HandleFunc("/v3/projects", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		log.Printf("%s %s", r.Method, r.URL.String())
		if r.Method == http.MethodGet {
			name := r.URL.Query().Get("name")
			list := []interface{}{}
			for _, p := range projects {
				if name != "" && p.Name != name {
					continue
				}
				list = append(list, map[string]interface{}{
					"id": p.ID, "name": p.Name, "domain_id": p.DomainID, "enabled": p.Enabled,
				})
			}
			err := json.NewEncoder(w).Encode(map[string]interface{}{"projects": list})
			if err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		if r.Method == http.MethodPost {
			var req struct {
				Project struct {
					Name     string `json:"name"`
					DomainID string `json:"domain_id"`
				} `json:"project"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid JSON body", http.StatusBadRequest)
				return
			}
			if req.Project.Name == "" || req.Project.DomainID == "" {
				http.Error(w, "Missing project name or domain_id", http.StatusBadRequest)
				return
			}
			id := RandomID("project")
			seq++
			p := &Project{ID: id, Name: req.Project.Name, DomainID: req.Project.DomainID, Enabled: true}
			projects[id] = p
			w.WriteHeader(http.StatusCreated)
			err := json.NewEncoder(w).Encode(map[string]interface{}{"project": p})
			if err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/v3/projects/", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		log.Printf("%s %s", r.Method, r.URL.String())
		pathParts := Parts(r.URL.Path)
		if len(pathParts) == 6 && pathParts[2] == "users" && pathParts[4] == "roles" && r.Method == http.MethodGet {
			projectID := pathParts[1]
			userID := pathParts[3]
			userRoles := []map[string]string{}
			if projMap, ok := roleAssignments[projectID]; ok {
				if roleIDs, ok := projMap[userID]; ok {
					for _, roleID := range roleIDs {
						if role, ok := roles[roleID]; ok {
							userRoles = append(userRoles, map[string]string{"id": role.ID, "name": role.Name})
						}
					}
				}
			}
			err := json.NewEncoder(w).Encode(map[string]interface{}{"roles": userRoles})
			if err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		if len(pathParts) == 7 && pathParts[2] == "users" && pathParts[4] == "roles" && r.Method == http.MethodPut {
			projectID := pathParts[1]
			userID := pathParts[3]
			roleID := pathParts[5]
			if _, ok := roleAssignments[projectID]; !ok {
				roleAssignments[projectID] = map[string][]string{}
			}
			roleAssignments[projectID][userID] = append(roleAssignments[projectID][userID], roleID)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if len(pathParts) == 3 {
			projectID := pathParts[2]
			p, ok := projects[projectID]
			if !ok {
				http.NotFound(w, r)
				return
			}
			if r.Method == http.MethodGet {
				err := json.NewEncoder(w).Encode(map[string]interface{}{"project": p})
				if err != nil {
					log.Printf("Error encoding JSON response: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				return
			}
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/v3/roles", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.String())
		switch r.Method {
		case http.MethodGet:
			name := r.URL.Query().Get("name")
			list := []interface{}{}
			for _, role := range roles {
				if name != "" && role.Name != name {
					continue
				}
				list = append(list, map[string]interface{}{"id": role.ID, "name": role.Name})
			}
			err := json.NewEncoder(w).Encode(map[string]interface{}{"roles": list})
			if err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		case http.MethodPost:
			var req struct {
				Role struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"role"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid JSON body", http.StatusBadRequest)
				return
			}
			if req.Role.Name == "" {
				http.Error(w, "Missing role id or name", http.StatusBadRequest)
				return
			}
			if req.Role.ID == "" {
				req.Role.ID = RandomID("role")
			}
			rl := &Role{ID: req.Role.ID, Name: req.Role.Name}
			roles[req.Role.ID] = rl
			w.WriteHeader(http.StatusCreated)
			err := json.NewEncoder(w).Encode(map[string]interface{}{"role": rl})
			if err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v3/roles/", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		log.Printf("%s %s", r.Method, r.URL.String())
		id := path.Base(r.URL.Path)
		rl, ok := roles[id]
		if !ok {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			err := json.NewEncoder(w).Encode(map[string]interface{}{
				"role": map[string]interface{}{"id": rl.ID, "name": rl.Name},
			})
			if err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		case http.MethodPatch, http.MethodPut:
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid JSON body", http.StatusBadRequest)
				return
			}
			if rReq, ok := req["role"].(map[string]interface{}); ok {
				if name, ok := rReq["name"].(string); ok {
					rl.Name = name
				}
			}
			err := json.NewEncoder(w).Encode(map[string]interface{}{
				"user": map[string]interface{}{"id": rl.ID, "name": rl.Name},
			})
			if err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		case http.MethodDelete:
			delete(roles, id)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/v3/users", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.String())
		mu.Lock()
		defer mu.Unlock()
		if r.Method == http.MethodDelete {
			part := Parts(r.URL.Path)
			userID := part[len(part)-1]
			delete(users, userID)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodGet {
			name := r.URL.Query().Get("name")
			list := []interface{}{}
			found := false
			for _, u := range users {
				if name != "" && u.Name != name {
					continue
				}
				found = true
				list = append(list, map[string]interface{}{
					"id": u.ID, "name": u.Name, "domain_id": u.DomainID,
					"default_project_id": u.DefaultProjectID, "enabled": u.Enabled,
					"federated": []interface{}{},
					"links": map[string]string{"self": baseURL + "/v3/users/" + u.ID},
					"password_expires_at": nil,
				})
			}
			if !found {
				for _, u := range users {
					if name != "" && u.ID != name {
						continue
					}
					list = append(list, map[string]interface{}{
						"id": u.ID, "name": u.Name, "domain_id": u.DomainID,
						"default_project_id": u.DefaultProjectID, "enabled": u.Enabled,
						"federated": []interface{}{},
						"links": map[string]string{"self": baseURL + "/v3/users/" + u.ID},
						"password_expires_at": nil,
					})
				}
			}
			err := json.NewEncoder(w).Encode(map[string]interface{}{"users": list})
			if err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		if r.Method == http.MethodPost {
			var req struct {
				User struct {
					Name     string `json:"name"`
					DomainID string `json:"domain_id"`
				} `json:"user"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid JSON body", http.StatusBadRequest)
				return
			}
			id := RandomID("user")
			seq++
			u := &User{ID: id, Name: req.User.Name, DomainID: req.User.DomainID, Enabled: true}
			users[id] = u
			w.WriteHeader(http.StatusCreated)
			err := json.NewEncoder(w).Encode(map[string]interface{}{
				"user": map[string]interface{}{
					"id": u.ID, "name": u.Name, "domain_id": u.DomainID, "enabled": true,
					"links": map[string]string{"self": baseURL + "/v3/users/" + u.ID},
				},
			})
			if err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
	})

	mux.HandleFunc("/v3/users/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.String())
		mu.Lock()
		defer mu.Unlock()
		id := path.Base(r.URL.Path)
		u, ok := users[id]
		if !ok {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			err := json.NewEncoder(w).Encode(map[string]interface{}{
				"user": map[string]interface{}{
					"id": u.ID, "name": u.Name, "domain_id": u.DomainID, "enabled": true,
					"links": map[string]string{"self": baseURL + "/v3/users/" + u.ID},
					"password_expires_at": "2099-01-01T00:00:00.000000Z",
					"federated": []interface{}{},
				},
			})
			if err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		case http.MethodPatch, http.MethodPut:
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid JSON body", http.StatusBadRequest)
				return
			}
			if userReq, ok := req["user"].(map[string]interface{}); ok {
				if name, ok := userReq["name"].(string); ok {
					u.Name = name
				}
			}
			err := json.NewEncoder(w).Encode(map[string]interface{}{
				"user": map[string]interface{}{"id": u.ID, "name": u.Name, "domain_id": u.DomainID, "enabled": true},
			})
			if err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		case http.MethodDelete:
			delete(users, id)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/v3/role_assignments", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.String())
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		q := r.URL.Query()
		log.Printf("role_assignments filters: user=%s group=%s project=%s domain=%s role=%s",
			q.Get("user.id"), q.Get("group.id"), q.Get("project.id"), q.Get("domain.id"), q.Get("role.id"))
		err := json.NewEncoder(w).Encode(map[string]interface{}{"role_assignments": []interface{}{}})
		if err != nil {
			log.Printf("Error encoding JSON response: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
}
