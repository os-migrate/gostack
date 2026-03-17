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

/* ============================
   Keystone
============================ */

// Domain represents a Keystone domain.
type Domain struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// User represents a Keystone user.
type User struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	DomainID         string            `json:"domain_id"`
	DefaultProjectID string            `json:"default_project_id,omitempty"`
	Enabled          bool              `json:"enabled"`
	Federated        []interface{}     `json:"federated"`
	Links            map[string]string `json:"links"`
	PasswordExpires  *string           `json:"password_expires_at"`
}

// Project represents a Keystone project.
type Project struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	DomainID string `json:"domain_id"`
	Enabled  bool   `json:"enabled"`
}

// Role represents a Keystone role.
type Role struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

/* ============================
   Nova
============================ */

// Server represents a Nova server.
type Server struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Status     string              `json:"status"`
	Flavor     map[string]string   `json:"flavor"`
	Image      map[string]string   `json:"image"`
	Networks   []map[string]string `json:"addresses"`
	Project    string              `json:"project_id"`
	AccessIPv4 string              `json:"accessIPv4,omitempty"`
}

/* ============================
   Cinder
============================ */

// Volume represents a Cinder volume.
type Volume struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Size        int                      `json:"size"`
	Status      string                   `json:"status"`
	Attachments []map[string]interface{} `json:"attachments"`
	Metadata    map[string]string        `json:"metadata"`
}
