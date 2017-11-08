package models

const (
	collectionID = "12345678"
)

var contacts = ContactDetails{
	Email:     "test@test.co.uk",
	Name:      "john test",
	Telephone: "01654 765432",
}

var methodology = GeneralDetails{
	Description: "some methodology description",
	HRef:        "http://localhost:22000//datasets/methodologies",
	Title:       "some methodology title",
}

var publications = GeneralDetails{
	Description: "some publication description",
	HRef:        "http://localhost:22000//datasets/publications",
	Title:       "some publication title",
}

var publisher = Publisher{
	Name: "The office of national statistics",
	Type: "government",
	HRef: "https://www.ons.gov.uk/",
}

var qmi = GeneralDetails{
	Description: "some qmi description",
	HRef:        "http://localhost:22000//datasets/123/qmi",
	Title:       "Quality and Methodology Information",
}

var relatedDatasets = GeneralDetails{
	HRef:  "http://localhost:22000//datasets/124",
	Title: "Census Age",
}

var inputDataset = Dataset{
	CollectionID: collectionID,
	Contacts: []ContactDetails{
		contacts,
	},
	Description: "census",
	Keywords:    []string{"test", "test2"},
	License:     "Office of National Statistics license",
	Methodologies: []GeneralDetails{
		methodology,
	},
	NationalStatistic: true,
	NextRelease:       "2016-05-05",
	Publications: []GeneralDetails{
		publications,
	},
	Publisher: &publisher,
	QMI:       &qmi,
	RelatedDatasets: []GeneralDetails{
		relatedDatasets,
	},
	ReleaseFrequency: "yearly",
	State:            "published",
	Theme:            "population",
	Title:            "CensusEthnicity",
	URI:              "http://localhost:22000/datasets/123/breadcrumbs",
}

var downloads = DownloadList{
	CSV: &DownloadObject{
		URL:  "https://www.aws/123",
		Size: "25mb",
	},
	XLS: &DownloadObject{
		URL:  "https://www.aws/1234",
		Size: "45mb",
	},
}

var links = VersionLinks{
	Dataset: &LinkObject{
		HRef: "http://localhost:22000/datasets/123",
		ID:   "3265vj48317tr4r34r3f",
	},
	Dimensions: &LinkObject{
		HRef: "http://localhost:22000/datasets/123/editions/2017/versions/1/dimensions",
	},
	Edition: &LinkObject{
		HRef: "http://localhost:22000/datasets/123/editions/2017",
		ID:   "asf87wafgu34gf87wfdgr",
	},
	Self: &LinkObject{
		HRef: "http://localhost:22000/datasets/123/editions/2017/versions/1",
	},
}

var createdVersion = Version{
	Downloads:   &downloads,
	Edition:     "2017",
	Links:       &links,
	ReleaseDate: "2016-04-04",
	State:       "created",
	Version:     1,
}

var associatedVersion = Version{
	CollectionID: collectionID,
	Downloads:    &downloads,
	Edition:      "2017",
	Links:        &links,
	ReleaseDate:  "2017-10-12",
	State:        "associated",
	Version:      1,
}

var publishedVersion = Version{
	CollectionID: collectionID,
	Downloads:    &downloads,
	Edition:      "2017",
	Links:        &links,
	ReleaseDate:  "2017-10-12",
	State:        "published",
	Version:      1,
}

var badInputData = struct {
	CollectionID int `json:"collection_id"`
}{
	CollectionID: 1,
}
