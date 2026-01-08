// Copyright 2024 The KitOps Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/google/licensecheck"
)

type fileType int

const (
	fileTypeModel fileType = iota
	fileTypeDataset
	fileTypeCode
	fileTypeDocs
	fileTypeMetadata
	fileTypePrompt
	fileTypeUnknown
)

var modelWeightsSuffixes = []string{
	".safetensors", ".pkl", ".joblib",
	// Pytorch suffixes
	".bin", ".pth", ".pt", ".mar", ".pt2", ".ptl",
	// Tensorflow
	".pb", ".ckpt", ".tflite", ".tfrecords",
	// NumPy
	".npy", ".npz",
	// Keras and others
	".keras", ".h5", ".caffemodel", ".pmml", ".coreml",
	// Other suffixes
	".gguf", ".ggml", ".ggmf", ".llamafile", ".onnx",
}

var docsSuffixes = []string{
	".md", ".adoc", ".html", ".pdf",
}

var metadataSuffixes = []string{
	".json", ".yaml", ".xml", ".txt",
}

var datasetSuffixes = []string{
	".tar", ".zip", ".parquet", ".csv",
}

// Files that are considered prompt files based on their names; entries should be in lowercase to support
// case-insensitive comparison.
var promptFilePatterns = []string{
	"agents.md",
	"skill.md",
	"claude.md",
}

// Generate a basic Kitfile by looking at the contents of a directory. Parameter
// packageOpt can be used to define metadata for the Kitfile (i.e. the package
// section), which is left empty if the parameter is nil.
func GenerateKitfile(dir *DirectoryListing, packageOpt *artifact.Package) (*artifact.KitFile, error) {
	output.Logf(output.LogLevelTrace, "Generating Kitfile in %s", dir.Path)
	kitfile := &artifact.KitFile{
		ManifestVersion: "1.0.0",
	}
	if packageOpt != nil {
		kitfile.Package = *packageOpt
	}

	// We can make sure all files are included by including a layer with path '.'
	// However, we only want to do this if it is necessary
	includeCatchallSection := false
	// Dirs we don't know how to handle automatically.
	var unprocessedDirPaths []string
	// Metadata files; we want these to be either model parts (if there is a model)
	// or datasets
	var modelFiles, metadataFiles []FileListing
	var detectedLicenseType string

	output.Logf(output.LogLevelTrace, "Reading directory contents")
	for _, file := range dir.Files {
		if constants.IsDefaultKitfileName(file.Name) {
			output.Logf(output.LogLevelTrace, "Skipping Kitfile '%s'", file.Name)
			// Skip Kitfile files (if present in the directory...). These won't be packed
			// either way.
			continue
		}

		// Check for "special" files (e.g. readme, license)
		if strings.HasPrefix(strings.ToLower(file.Name), "readme") {
			output.Logf(output.LogLevelTrace, "Found readme file '%s'", file.Name)
			kitfile.Docs = append(kitfile.Docs, artifact.Docs{
				Path:        file.Name,
				Description: "Readme file",
			})
			continue
		} else if strings.HasPrefix(strings.ToLower(file.Name), "license") {
			output.Logf(output.LogLevelTrace, "Found license file '%s'", file.Name)
			kitfile.Docs = append(kitfile.Docs, artifact.Docs{
				Path:        file.Name,
				Description: "License file",
			})
			licenseType, err := detectLicense(file.Path)
			if err != nil {
				output.Debugf("Error determining license type: %s", err)
				output.Logf(output.LogLevelWarn, "Unable to determine license type")
			}
			detectedLicenseType = licenseType
			output.Logf(output.LogLevelTrace, "Detected license %s for license file", detectedLicenseType)
			continue
		}

		// Try to determine type based on file extension
		// To support multi-part models, we need to collect all paths and decide
		// which one is the model and which one(s) are parts
		switch determineFileType(file.Path) {
		case fileTypeModel:
			modelFiles = append(modelFiles, file)
		case fileTypeMetadata:
			// Metadata should be included in either Model or Datasets, depending on
			// other contents
			output.Logf(output.LogLevelTrace, "Detected metadata file '%s'", file.Path)
			metadataFiles = append(metadataFiles, file)
		case fileTypeDocs:
			kitfile.Docs = append(kitfile.Docs, artifact.Docs{Path: file.Path})
		case fileTypeDataset:
			kitfile.DataSets = append(kitfile.DataSets, artifact.DataSet{Path: file.Path})
		case fileTypePrompt:
			kitfile.Prompts = append(kitfile.Prompts, artifact.Prompt{Path: file.Path})
		default:
			output.Logf(output.LogLevelTrace, "File %s is either code or unknown type. Will be added as a catch-all section", file.Path)
			// File is either code or unknown; we'll have to include it in a catch-all section
			includeCatchallSection = true
		}
	}

	for _, subDir := range dir.Subdirs {
		dirModelFiles, err := addDirToKitfile(kitfile, subDir)
		if err != nil {
			output.Logf(output.LogLevelTrace, "Failed to determine type for directory %s: %s", subDir.Path, err)
			unprocessedDirPaths = append(unprocessedDirPaths, subDir.Path)
		}
		modelFiles = append(modelFiles, dirModelFiles...)
		continue
	}

	if len(modelFiles) > 0 {
		if err := addModelToKitfile(kitfile, modelFiles); err != nil {
			return nil, fmt.Errorf("failed to add model to Kitfile: %w", err)
		}
		output.Logf(output.LogLevelTrace, "Adding metadata files as model parts")
		for _, metadataFile := range metadataFiles {
			kitfile.Model.Parts = append(kitfile.Model.Parts, artifact.ModelPart{Path: metadataFile.Path})
		}
	} else {
		output.Logf(output.LogLevelTrace, "No model detected; adding metadata files as dataset layers")
		for _, metadataFile := range metadataFiles {
			kitfile.DataSets = append(kitfile.DataSets, artifact.DataSet{Path: metadataFile.Path})
		}
	}

	// Decide how to handle remaining paths. Either package them in one large code layer with basePath
	// or as separate layers for each directory.
	output.Logf(output.LogLevelTrace, "Unable to process %d paths in %s", len(unprocessedDirPaths), dir.Path)
	if includeCatchallSection || len(unprocessedDirPaths) > 5 {
		output.Logf(output.LogLevelTrace, "Adding catch-all code layer to include files in %s", dir.Path)
		// Overwrite any code layers we added before; this is cleaner than e.g. having a layer for '.' and a layer for 'src'
		kitfile.Code = []artifact.Code{{Path: "."}}
	} else {
		for _, path := range unprocessedDirPaths {
			kitfile.Code = append(kitfile.Code, artifact.Code{Path: path})
		}
	}

	// If we detected a license, try to attach it to the Kitfile section that makes sense
	if kitfile.Model != nil && detectedLicenseType != "" {
		kitfile.Model.License = detectedLicenseType
	} else if len(kitfile.DataSets) == 1 {
		kitfile.DataSets[0].License = detectedLicenseType
	} else if len(kitfile.Code) == 1 {
		kitfile.Code[0].License = detectedLicenseType
	} else {
		output.Logf(output.LogLevelTrace, "Unsure what license applies to, adding to Kitfile package")
		kitfile.Package.License = detectedLicenseType
	}

	return kitfile, nil
}

func addDirToKitfile(kitfile *artifact.KitFile, dir DirectoryListing) (modelFiles []FileListing, err error) {
	switch dir.Name {
	case "docs":
		output.Logf(output.LogLevelTrace, "Directory %s interpreted as documentation", dir.Name)
		kitfile.Docs = append(kitfile.Docs, artifact.Docs{
			Path: dir.Path,
		})
		return nil, nil
	case "src", "pkg", "lib", "build":
		output.Logf(output.LogLevelTrace, "Directory %s interpreted as code", dir.Name)
		kitfile.Code = append(kitfile.Code, artifact.Code{
			Path: dir.Path,
		})
		return nil, nil
	}

	// Sort entries in the directory to try and figure out what it contains. We'll reuse the
	// fact that the fileTypes are enumerated using iota (and so are ints) to index correctly.
	// Avoid using maps here since they iterate in a random order.
	directoryContents := [int(fileTypeUnknown) + 1][]string{}
	for _, subdir := range dir.Subdirs {
		// We can, in the future, recurse deeper into the directory tree here. For now, treat secondary dirs as unknowns
		directoryContents[int(fileTypeUnknown)] = append(directoryContents[int(fileTypeUnknown)], subdir.Path)
	}

	var metadataFiles []FileListing
	for _, file := range dir.Files {
		fileType := determineFileType(file.Name)
		if fileType == fileTypeModel {
			modelFiles = append(modelFiles, file)
		}
		if fileType == fileTypeMetadata {
			metadataFiles = append(metadataFiles, file)
		}
		directoryContents[int(fileType)] = append(directoryContents[int(fileType)], file.Path)
	}

	// Try to detect directories that contain e.g. only datasets so we can add them
	overallFiletype := fileTypeUnknown
	directoryHasMixedContents := false
	for fType, files := range directoryContents {
		if len(files) > 0 && fileType(fType) != fileTypeMetadata {
			if overallFiletype != fileTypeUnknown {
				output.Logf(output.LogLevelTrace, "Detected mixed contents within directory %s", dir.Path)
				directoryHasMixedContents = true
			}
			overallFiletype = fileType(fType)
		}
	}
	if directoryHasMixedContents {
		return modelFiles, fmt.Errorf("mixed content in directory; unable to determine type")
	}
	switch overallFiletype {
	case fileTypeModel:
		output.Logf(output.LogLevelTrace, "Interpreting directory %s as a model directory", dir.Path)
		// Include any metadata files as modelParts later
		modelFiles = append(modelFiles, metadataFiles...)
	case fileTypeDataset:
		output.Logf(output.LogLevelTrace, "Interpreting directory %s as a dataset directory", dir.Path)
		kitfile.DataSets = append(kitfile.DataSets, artifact.DataSet{Path: dir.Path})
	case fileTypeDocs:
		output.Logf(output.LogLevelTrace, "Interpreting directory %s as a docs directory", dir.Path)
		kitfile.Docs = append(kitfile.Docs, artifact.Docs{Path: dir.Path})
	case fileTypePrompt:
		output.Logf(output.LogLevelTrace, "Interpreting directory %s as a prompts directory", dir.Path)
		kitfile.Prompts = append(kitfile.Prompts, artifact.Prompt{Path: dir.Path})
	default:
		output.Logf(output.LogLevelTrace, "Could not determine type for directory %s", dir.Path)
		// If it's overall code, metadata, or unknown, just return it as unprocessed and let it be added as a Code section
		// later
		return modelFiles, fmt.Errorf("directory should be handled as Code")
	}

	return modelFiles, nil
}

func determineFileType(filename string) fileType {
	// Check for prompt file matches: either exact name or containing '.prompt'. This has to come first
	// as the prompt may have colliding suffixes.
	baseName := strings.ToLower(filepath.Base(filename))
	if slices.Contains(promptFilePatterns, baseName) || strings.Contains(baseName, ".prompt") {
		return fileTypePrompt
	}

	if anySuffix(filename, modelWeightsSuffixes) {
		return fileTypeModel
	}
	// Metadata should be included in either Model or Datasets, depending on
	// other contents
	if anySuffix(filename, metadataSuffixes) {
		return fileTypeMetadata
	}
	if anySuffix(filename, docsSuffixes) {
		return fileTypeDocs
	}
	if anySuffix(filename, datasetSuffixes) {
		return fileTypeDataset
	}
	return fileTypeUnknown

}

func addModelToKitfile(kitfile *artifact.KitFile, files []FileListing) error {
	if len(files) == 0 {
		return nil
	}

	if len(files) == 1 {
		file := files[0]
		kitfile.Model = &artifact.Model{
			Path: file.Path,
			Name: strings.TrimSuffix(file.Path, filepath.Ext(file.Name)),
		}
		return nil
	}

	// We'll handle two cases here: 1) the Model is split into multiple files (e.g. safetensors) or 2) there is a
	// main model plus smaller adaptor(s)
	var largestFile FileListing
	largestSize := int64(0)
	averageSize := int64(0)
	for _, modelFile := range files {
		if modelFile.Size > largestSize {
			largestSize = modelFile.Size
			largestFile = modelFile
		}
		averageSize = averageSize + modelFile.Size
	}
	// Integer division is probably fine here; at most we're off by a byte.
	averageSize = averageSize / int64(len(files))

	// If the biggest file is 1.5x the average, make it the model and the rest parts; otherwise, add
	// all parts in lexical order
	if largestSize > averageSize+(averageSize/2) {
		kitfile.Model = &artifact.Model{
			Path: largestFile.Path,
		}
		kitfile.Model.Name = strings.TrimSuffix(largestFile.Path, filepath.Ext(largestFile.Name))
		for _, file := range files {
			if file == largestFile {
				continue
			}
			kitfile.Model.Parts = append(kitfile.Model.Parts, artifact.ModelPart{
				Path: file.Path,
			})
		}
	} else {
		kitfile.Model = &artifact.Model{
			Path: files[0].Path,
		}
		for _, file := range files[1:] {
			kitfile.Model.Parts = append(kitfile.Model.Parts, artifact.ModelPart{
				Path: file.Path,
			})
		}
	}
	return nil
}

func detectLicense(licensePath string) (string, error) {
	license, err := os.ReadFile(licensePath)
	if err != nil {
		return "", fmt.Errorf("failed to read license file: %w", err)
	}
	cov := licensecheck.Scan(license)
	if len(cov.Match) == 1 {
		return cov.Match[0].ID, nil
	} else {
		return "", fmt.Errorf("multiple licenses matched license file")
	}
}

func anySuffix(query string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(query, suffix) {
			return true
		}
	}
	return false
}
