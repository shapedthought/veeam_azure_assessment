package main

type SessionInfo struct {
	Offset     int `json:"offset"`
	Limit      int `json:"limit"`
	TotalCount int `json:"totalCount"`
	Results    []struct {
		ID                 string `json:"id"`
		Type               string `json:"type"`
		LocalizedType      string `json:"localizedType"`
		ExecutionStartTime string `json:"executionStartTime"`
		ExecutionStopTime  string `json:"executionStopTime"`
		ExecutionDuration  string `json:"executionDuration"`
		Status             string `json:"status"`
		Usn                int    `json:"usn"`
		BackupJobInfo      struct {
			PolicyID                string `json:"policyId"`
			PolicyName              string `json:"policyName"`
			ProtectedInstancesCount int    `json:"protectedInstancesCount"`
			PolicyRemoved           bool   `json:"policyRemoved"`
		} `json:"backupJobInfo,omitempty"`
		Links struct {
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
			Log struct {
				Href string `json:"href"`
			} `json:"log"`
		} `json:"_links"`
		EmbeddedResources struct {
		} `json:"_embeddedResources"`
		RepositoryJobInfo struct {
			RepositoryID      string `json:"repositoryId"`
			RepositoryName    string `json:"repositoryName"`
			RepositoryRemoved bool   `json:"repositoryRemoved"`
		} `json:"repositoryJobInfo,omitempty"`
	} `json:"results"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
}

type SessionLog struct {
	JobSessionID string `json:"jobSessionId"`
	Log          []struct {
		LogTime            string `json:"logTime"`
		Status             string `json:"status"`
		Message            string `json:"message"`
		ExecutionStartTime string `json:"executionStartTime"`
		ExecutionDuration  string `json:"executionDuration"`
		ResourceHashID     string `json:"resourceHashId,omitempty"`
	} `json:"log"`
}
