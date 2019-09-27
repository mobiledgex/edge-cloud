package k8smgmt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var testCustomizationFileList = []string{"file1.yml", "file2.yml", "file3.yml"}
var testInvalidBooleanAnnotationStr = "opt1=val1.2,opt2"
var testValidBooleanAnnotationStr = "opt1=val1.2,opt2=true"

var testHemlExistingRepoChart = "existing/testchart"
var testHemlNewRepoChart = "testcharts/testchart"
var testHelmNewRepoImage = "http://testchartRepo.mex/charts:testcharts/testchart"
var testInvalidHelmImageWithPort = "http://testchartRepo.mex/charts:8080:testcharts/testchart"
var testInvalidHelmImageNoRepoName = "http://testchartRepo.mex/charts:testchart"

func TestHelm(t *testing.T) {
	str := getHelmYamlOpt(testCustomizationFileList)
	require.Equal(t, "-f file1.yml,file2.yml,file3.yml", str)
	str = getHelmYamlOpt([]string{})
	require.Equal(t, "", str)

	str, err := getHelmInstallOptsString("invalid annotations string")
	require.NotNil(t, err, "This should return an error")
	require.Equal(t, "", str, "error should return an empty string")
	str, err = getHelmInstallOptsString("")
	require.Nil(t, err, "No annotations should be a valid string")
	require.Equal(t, "", str, "empty options for empty annotations")
	str, err = getHelmInstallOptsString(testInvalidBooleanAnnotationStr)
	require.NotNil(t, err, "Incorrect way of specifying boolean option")
	require.Contains(t, err.Error(), "Invalid annotations stirng")
	require.Equal(t, "", str, "error should return an empty string")
	str, err = getHelmInstallOptsString(testValidBooleanAnnotationStr)
	require.Nil(t, err)
	require.Equal(t, "--opt1 val1.2 --opt2", str, "Invalid options string returned")

	repo, chart, err := getHelmRepoAndChart(testHemlExistingRepoChart)
	require.Nil(t, err)
	require.Equal(t, "", repo, "Repo should be empty")
	require.Equal(t, testHemlExistingRepoChart, chart)
	repo, chart, err = getHelmRepoAndChart(testHelmNewRepoImage)
	require.Nil(t, err)
	require.Equal(t, "testcharts http://testchartRepo.mex/charts", repo, "Couldn't get repo from path")
	require.Equal(t, testHemlNewRepoChart, chart)
	repo, chart, err = getHelmRepoAndChart(testInvalidHelmImageWithPort)
	require.NotNil(t, err, "Port number in URL is not supported")
	require.Contains(t, err.Error(), "Could not parse the repository")
	require.Equal(t, "", repo, "Error should result in an empty repo")
	require.Equal(t, "", chart, "Error should result in an emty chart")
	repo, chart, err = getHelmRepoAndChart(testInvalidHelmImageNoRepoName)
	require.NotNil(t, err, "Repo should have a name")
	require.Contains(t, err.Error(), "Could not parse the chart")
	require.Equal(t, "", repo, "Error should result in an empty repo")
	require.Equal(t, "", chart, "Error should result in an emty chart")
	//Random string test
	repo, chart, err = getHelmRepoAndChart("random string : ")
	require.NotNil(t, err, "Random string should be invalid path")
	require.Contains(t, err.Error(), "Could not parse the chart")
	require.Equal(t, "", repo, "Error should result in an empty repo")
	require.Equal(t, "", chart, "Error should result in an emty chart")

}
