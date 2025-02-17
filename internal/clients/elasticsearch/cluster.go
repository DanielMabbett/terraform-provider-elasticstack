package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func PutSnapshotRepository(ctx context.Context, apiClient *clients.ApiClient, repository *models.SnapshotRepository) diag.Diagnostics {
	var diags diag.Diagnostics
	snapRepoBytes, err := json.Marshal(repository)
	if err != nil {
		return diag.FromErr(err)
	}
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Snapshot.CreateRepository(repository.Name, bytes.NewReader(snapRepoBytes), esClient.Snapshot.CreateRepository.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update the snapshot repository"); diags.HasError() {
		return diags
	}

	return diags
}

func GetSnapshotRepository(ctx context.Context, apiClient *clients.ApiClient, name string) (*models.SnapshotRepository, diag.Diagnostics) {
	var diags diag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	req := esClient.Snapshot.GetRepository.WithRepository(name)
	res, err := esClient.Snapshot.GetRepository(req, esClient.Snapshot.GetRepository.WithContext(ctx))
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to get the information about snapshot repository: %s", name)); diags.HasError() {
		return nil, diags
	}
	snapRepoResponse := make(map[string]models.SnapshotRepository)
	if err := json.NewDecoder(res.Body).Decode(&snapRepoResponse); err != nil {
		return nil, diag.FromErr(err)
	}

	if currentRepo, ok := snapRepoResponse[name]; ok {
		if len(currentRepo.Name) <= 0 {
			currentRepo.Name = name
		}
		return &currentRepo, diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to find requested repository",
		Detail:   fmt.Sprintf(`Repository "%s" is missing in the ES API response`, name),
	})
	return nil, diags
}

func DeleteSnapshotRepository(ctx context.Context, apiClient *clients.ApiClient, name string) diag.Diagnostics {
	var diags diag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Snapshot.DeleteRepository([]string{name}, esClient.Snapshot.DeleteRepository.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete snapshot repository: %s", name)); diags.HasError() {
		return diags
	}
	return diags
}

func PutSlm(ctx context.Context, apiClient *clients.ApiClient, slm *models.SnapshotPolicy) diag.Diagnostics {
	var diags diag.Diagnostics

	slmBytes, err := json.Marshal(slm)
	if err != nil {
		return diag.FromErr(err)
	}
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	req := esClient.SlmPutLifecycle.WithBody(bytes.NewReader(slmBytes))
	res, err := esClient.SlmPutLifecycle(slm.Id, req, esClient.SlmPutLifecycle.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update the SLM"); diags.HasError() {
		return diags
	}

	return diags
}

func GetSlm(ctx context.Context, apiClient *clients.ApiClient, slmName string) (*models.SnapshotPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	req := esClient.SlmGetLifecycle.WithPolicyID(slmName)
	res, err := esClient.SlmGetLifecycle(req, esClient.SlmGetLifecycle.WithContext(ctx))
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if diags := utils.CheckError(res, "Unable to get SLM policy from ES API"); diags.HasError() {
		return nil, diags
	}
	type SlmResponse = map[string]struct {
		Policy models.SnapshotPolicy `json:"policy"`
	}
	var slmResponse SlmResponse
	if err := json.NewDecoder(res.Body).Decode(&slmResponse); err != nil {
		return nil, diag.FromErr(err)
	}
	if slm, ok := slmResponse[slmName]; ok {
		return &slm.Policy, diags
	}
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to find the SLM policy in the response",
		Detail:   fmt.Sprintf(`Unable to find "%s" policy in the ES API response.`, slmName),
	})
	return nil, diags
}

func DeleteSlm(ctx context.Context, apiClient *clients.ApiClient, slmName string) diag.Diagnostics {
	var diags diag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.SlmDeleteLifecycle(slmName, esClient.SlmDeleteLifecycle.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete SLM policy: %s", slmName)); diags.HasError() {
		return diags
	}

	return diags
}

func PutSettings(ctx context.Context, apiClient *clients.ApiClient, settings map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	settingsBytes, err := json.Marshal(settings)
	if err != nil {
		diag.FromErr(err)
	}
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Cluster.PutSettings(bytes.NewReader(settingsBytes), esClient.Cluster.PutSettings.WithContext(ctx))
	if err != nil {
		diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to update cluster settings."); diags.HasError() {
		return diags
	}
	return diags
}

func GetSettings(ctx context.Context, apiClient *clients.ApiClient) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	req := esClient.Cluster.GetSettings.WithFlatSettings(true)
	res, err := esClient.Cluster.GetSettings(req, esClient.Cluster.GetSettings.WithContext(ctx))
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to read cluster settings."); diags.HasError() {
		return nil, diags
	}

	clusterSettings := make(map[string]interface{})
	if err := json.NewDecoder(res.Body).Decode(&clusterSettings); err != nil {
		return nil, diag.FromErr(err)
	}
	return clusterSettings, diags
}

func GetScript(ctx context.Context, apiClient *clients.ApiClient, id string) (*models.Script, diag.Diagnostics) {
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	res, err := esClient.GetScript(id, esClient.GetScript.WithContext(ctx))
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to get stored script: %s", id)); diags.HasError() {
		return nil, diags
	}
	var scriptResponse struct {
		Script *models.Script `json:"script"`
	}
	if err := json.NewDecoder(res.Body).Decode(&scriptResponse); err != nil {
		return nil, diag.FromErr(err)
	}

	return scriptResponse.Script, nil
}

func PutScript(ctx context.Context, apiClient *clients.ApiClient, script *models.Script) diag.Diagnostics {
	req := struct {
		Script *models.Script `json:"script"`
	}{
		script,
	}
	scriptBytes, err := json.Marshal(req)
	if err != nil {
		return diag.FromErr(err)
	}
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.PutScript(script.ID, bytes.NewReader(scriptBytes), esClient.PutScript.WithContext(ctx), esClient.PutScript.WithScriptContext(script.Context))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to put stored script"); diags.HasError() {
		return diags
	}
	return nil
}

func DeleteScript(ctx context.Context, apiClient *clients.ApiClient, id string) diag.Diagnostics {
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.DeleteScript(id, esClient.DeleteScript.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete script: %s", id)); diags.HasError() {
		return diags
	}
	return nil
}
