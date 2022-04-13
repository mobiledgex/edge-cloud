// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8smgmt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var testCustomizationFileList = []string{"file1.yml", "file2.yml", "file3.yml"}
var testInvalidBooleanAnnotationStr = "version=val1.2,timeout"
var testValidBooleanAnnotationStr = "version=val1.2,wait=true,timeout=60"
var testInvalidAnnotationsVal = "version=1.2.2;touch /tmp/broken"
var testInvalidAnnotationsOpt = "version`touch /tmp/broken`;ls -lf /tmp/broken=1.2.2"

var testHemlExistingRepoChart = "existing/testchart"
var testHemlNewRepoChart = "testcharts/testchart"
var testHelmNewRepoImage = "http://testchartRepo.mex/charts:testcharts/testchart"
var testHelmImageWithPort = "http://testchartRepo.mex:8000/charts:testcharts/testchart"
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
	require.Contains(t, err.Error(), "Invalid annotations string")
	require.Equal(t, "", str, "error should return an empty string")
	str, err = getHelmInstallOptsString(testInvalidAnnotationsVal)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\";\" not allowed in annotations")
	require.Equal(t, "", str, "error should return an empty string")
	str, err = getHelmInstallOptsString(testInvalidAnnotationsOpt)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"`\" not allowed in annotations")
	require.Equal(t, "", str, "error should return an empty string")

	str, err = getHelmInstallOptsString(testValidBooleanAnnotationStr)
	require.Nil(t, err)
	require.Equal(t, "--version \"val1.2\" --wait --timeout 60", str, "Invalid options string returned")

	repo, chart, err := getHelmRepoAndChart(testHemlExistingRepoChart)
	require.Nil(t, err)
	require.Equal(t, "", repo, "Repo should be empty")
	require.Equal(t, testHemlExistingRepoChart, chart)
	repo, chart, err = getHelmRepoAndChart(testHelmNewRepoImage)
	require.Nil(t, err)
	require.Equal(t, "testcharts http://testchartRepo.mex/charts", repo, "Couldn't get repo from path")
	require.Equal(t, testHemlNewRepoChart, chart)
	repo, chart, err = getHelmRepoAndChart(testHelmImageWithPort)
	require.Nil(t, err)
	require.Equal(t, "testcharts http://testchartRepo.mex:8000/charts", repo)
	require.Equal(t, testHemlNewRepoChart, chart)
	repo, chart, err = getHelmRepoAndChart(testInvalidHelmImageNoRepoName)
	require.NotNil(t, err, "Repo should have a name")
	require.Contains(t, err.Error(), "Could not parse the chart")
	require.Equal(t, "", repo, "Error should result in an empty repo")
	require.Equal(t, "", chart, "Error should result in an emty chart")
	//Random string test
	repo, chart, err = getHelmRepoAndChart("random string : ")
	require.NotNil(t, err, "Random string should be invalid path")
	require.Equal(t, "", repo, "Error should result in an empty repo")
	require.Equal(t, "", chart, "Error should result in an emty chart")

}
