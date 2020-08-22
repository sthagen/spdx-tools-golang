// SPDX-License-Identifier: Apache-2.0 OR GPL-2.0-or-later

package parser2v2

import (
	"fmt"
	gordfParser "github.com/RishabhBhatnagar/gordf/rdfloader/parser"
	"github.com/spdx/tools-golang/spdx"
	"strings"
)

// returns a file instance and the error if any encountered.
func (parser *rdfParser2_2) getFileFromNode(fileNode *gordfParser.Node) (file *spdx.File2_2, err error) {
	file = &spdx.File2_2{}

	err = setFileIdentifier(fileNode.ID, file, parser) // 4.2
	if err != nil {
		return nil, err
	}

	for _, subTriple := range parser.nodeToTriples(fileNode) {
		switch subTriple.Predicate.ID {
		case SPDX_FILE_NAME: // 4.1
			// cardinality: exactly 1
			file.FileName = subTriple.Object.ID
		case SPDX_NAME:
			// cardinality: exactly 1
			// todo: check where it will be set in the golang-tools spdx-data-model
		case RDF_TYPE:
			// cardinality: exactly 1
		case SPDX_FILE_TYPE: // 4.3
			// cardinality: min 0
			fileType := ""
			fileType, err = parser.getFileTypeFromUri(subTriple.Object.ID)
			file.FileType = append(file.FileType, fileType)
		case SPDX_CHECKSUM: // 4.4
			// cardinality: min 1
			err = parser.setFileChecksumFromNode(file, subTriple.Object)
		case SPDX_LICENSE_CONCLUDED: // 4.5
			// cardinality: (exactly 1 anyLicenseInfo) or (None) or (Noassertion)
			anyLicense, err := parser.getAnyLicenseFromNode(subTriple.Object)
			if err != nil {
				return nil, fmt.Errorf("error parsing licenseConcluded: %v", err)
			}
			file.LicenseConcluded = anyLicense.ToLicenseString()
		case SPDX_LICENSE_INFO_IN_FILE: // 4.6
			// cardinality: min 1
			lastPart := getLastPartOfURI(subTriple.Object.ID)
			file.LicenseInfoInFile = append(file.LicenseInfoInFile, lastPart)
		case SPDX_LICENSE_COMMENTS: // 4.7
			// cardinality: max 1
			file.LicenseComments = subTriple.Object.ID
		case SPDX_COPYRIGHT_TEXT: // 4.8
			// cardinality: exactly 1
			file.FileCopyrightText = subTriple.Object.ID
		case SPDX_LICENSE_INFO_FROM_FILES:
			// todo: implement it. It is not defined in the tools-golang model.
		// deprecated artifactOf (see sections 4.9, 4.10, 4.11)
		case SPDX_ARTIFACT_OF:
			// cardinality: min 0
			var artifactOf *spdx.ArtifactOfProject2_2
			artifactOf, err = parser.getArtifactFromNode(subTriple.Object)
			file.ArtifactOfProjects = append(file.ArtifactOfProjects, artifactOf)
		case RDFS_COMMENT: // 4.12
			// cardinality: max 1
			file.FileComment = subTriple.Object.ID
		case SPDX_NOTICE_TEXT: // 4.13
			// cardinality: max 1
			file.FileNotice = subTriple.Object.ID
		case SPDX_FILE_CONTRIBUTOR: // 4.14
			// cardinality: min 0
			file.FileContributor = append(file.FileContributor, subTriple.Object.ID)
		case SPDX_FILE_DEPENDENCY:
			// cardinality: min 0
			file, err := parser.getFileFromNode(subTriple.Object)
			if err != nil {
				return nil, fmt.Errorf("error setting a file dependency in a file: %v", err)
			}
			parser.files[file.FileSPDXIdentifier] = file
		case SPDX_ATTRIBUTION_TEXT:
			// cardinality: min 0
			file.FileAttributionTexts = append(file.FileAttributionTexts, subTriple.Object.ID)
		case SPDX_ANNOTATION: // unknown section
			err = parser.parseAnnotationFromNode(subTriple.Object)
		case SPDX_RELATIONSHIP: // unknown section
			err = parser.parseRelationship(subTriple)
		default:
			return nil, fmt.Errorf("unknown triple predicate id %s", subTriple.Predicate.ID)
		}
		if err != nil {
			return nil, err
		}
	}
	return file, nil
}

func (parser *rdfParser2_2) setFileChecksumFromNode(file *spdx.File2_2, checksumNode *gordfParser.Node) error {
	checksumAlgorithm, checksumValue, err := parser.getChecksumFromNode(checksumNode)
	if err != nil {
		return nil
	}
	switch checksumAlgorithm {
	case "MD5":
		file.FileChecksumMD5 = checksumValue
	case "SHA1":
		file.FileChecksumSHA1 = checksumValue
	case "SHA256":
		file.FileChecksumSHA256 = checksumValue
	case "":
		return fmt.Errorf("empty checksum algorithm and value")
	default:
		return fmt.Errorf("unknown checksumAlgorithm %s while parsing a file", checksumAlgorithm)
	}
	return nil
}

func (parser *rdfParser2_2) getArtifactFromNode(node *gordfParser.Node) (*spdx.ArtifactOfProject2_2, error) {
	artifactOf := &spdx.ArtifactOfProject2_2{}
	// setting artifactOfProjectURI attribute (which is optional)
	if node.NodeType == gordfParser.IRI {
		artifactOf.URI = node.ID
	}
	// parsing rest triples and attributes of the artifact.
	for _, triple := range parser.nodeToTriples(node) {
		switch triple.Predicate.ID {
		case RDF_TYPE:
		case DOAP_HOMEPAGE:
			artifactOf.HomePage = triple.Object.ID
		case DOAP_NAME:
			artifactOf.Name = triple.Object.ID
		default:
			return nil, fmt.Errorf("error parsing artifactOf predicate %s", triple.Predicate.ID)
		}
	}
	return artifactOf, nil
}

func (parser *rdfParser2_2) getFileTypeFromUri(uri string) (string, error) {
	// fileType is given as a uri. for example: http://spdx.org/rdf/terms#fileType_text
	lastPart := getLastPartOfURI(uri)
	if !strings.HasPrefix(lastPart, "fileType_") {
		return "", fmt.Errorf("fileType Uri must begin with fileTYpe_. found: %s", lastPart)
	}
	return strings.TrimPrefix(lastPart, "fileType_"), nil
}

// populates parser.doc.UnpackagedFiles by a list of files which are not
// associated with a package by the hasFile attribute
// assumes: all the packages are already parsed.
func (parser *rdfParser2_2) setUnpackagedFiles() {
	for fileID := range parser.files {
		if !parser.assocWithPackage[fileID] {
			parser.doc.UnpackagedFiles[fileID] = parser.files[fileID]
		}
	}
}

func setFileIdentifier(idURI string, file *spdx.File2_2, parser *rdfParser2_2) (err error) {
	idURI = strings.TrimSpace(idURI)
	uriFragment := getLastPartOfURI(idURI)
	file.FileSPDXIdentifier, err = ExtractElementID(uriFragment)
	if err != nil {
		return fmt.Errorf("error setting file identifier: %s", err)
	}
	return nil
}
