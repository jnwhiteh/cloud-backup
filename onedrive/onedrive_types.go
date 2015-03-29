package main

import "time"


type AsyncOperationStatus struct {
	Status string `json:"status"` // notStarted | inProgress | completed | updating | failed | deletePending | deleteFailed | waiting
	Operation string `json:"operation"`
	PercentageComplete float64 `json:"percentageComplete"`
}

type Audio struct {
	Composers string `json:"composers"`
	Disc float64 `json:"disc"`
	Track float64 `json:"track"`
	Year float64 `json:"year"`
	Album string `json:"album"`
	Copyright string `json:"copyright"`
	Genre string `json:"genre"`
	IsVariableBitrate bool `json:"isVariableBitrate"`
	Artist string `json:"artist"`
	Bitrate float64 `json:"bitrate"`
	TrackCount float64 `json:"trackCount"`
	AlbumArtist string `json:"albumArtist"`
	DiscCount float64 `json:"discCount"`
	Duration float64 `json:"duration"`
	HasDrm bool `json:"hasDrm"`
	Title string `json:"title"`
}

type Deleted struct {
}

type Drive struct {
	Id string `json:"id"`
	DriveType string `json:"driveType"`
	Owner *IdentitySet `json:"owner"`
	Quota *Quota `json:"quota"`
}

type File struct {
	Hashes *Hashes `json:"hashes"`
	MimeType string `json:"mimeType"`
}

type Folder struct {
	ChildCount float64 `json:"childCount"`
}

type Hashes struct {
	Crc32Hash string `json:"crc32Hash"` // hex
	Sha1Hash string `json:"sha1Hash"` // hex
}

type Identity struct {
	DisplayName string `json:"displayName"` // optional
	Id string `json:"id"`
}

type IdentitySet struct {
	User *Identity `json:"user"`
	Application *Identity `json:"application"`
	Device *Identity `json:"device"`
}

type Image struct {
	Width float64 `json:"width"`
	Height float64 `json:"height"`
}

type Item struct {
	LastModifiedBy *IdentitySet `json:"lastModifiedBy"`
	WebUrl string `json:"webUrl"` // url
	Photo *Photo `json:"photo"`
	Location *Location `json:"location"`
	Name string `json:"name"`
	LastModifiedDateTime time.Time `json:"lastModifiedDateTime"` // string timestamp
	File *File `json:"file"`
	Video *Video `json:"video"`
	Instancecontent_sourceUrl string `json:"@content.sourceUrl"` // url
	CTag string `json:"cTag"` // etag
	Size float64 `json:"size"`
	Folder *Folder `json:"folder"`
	Image *Image `json:"image"`
	Deleted *Deleted `json:"deleted"`
	SpecialFolder *SpecialFolder `json:"specialFolder"`
	Thumbnails []*ThumbnailSet `json:"thumbnails"`
	Id string `json:"id"` // identifier
	ETag string `json:"eTag"` // etag
	CreatedBy *IdentitySet `json:"createdBy"`
	CreatedDateTime time.Time `json:"createdDateTime"` // string timestamp
	ParentReference *ItemReference `json:"parentReference"`
	Children []*Item `json:"children"`
	Audio *Audio `json:"audio"`
	Instancename_conflictBehavior string `json:"@name.conflictBehavior"`
	Instancecontent_downloadUrl string `json:"@content.downloadUrl"` // url
}

type ItemReference struct {
	DriveId string `json:"driveId"` // identifier
	Id string `json:"id"` // identifier
	Path string `json:"path"` // path
}

type SpecialFolder struct {
	Name string `json:"name"`
}

type Location struct {
	Latitude float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude float64 `json:"altitude"`
}

type Permission struct {
	Id string `json:"id"`
	Roles []string `json:"roles"` // read|write
	Link *SharingLink `json:"link"`
	InheritedFrom *ItemReference `json:"inheritedFrom"`
}

type Photo struct {
	CameraMake string `json:"cameraMake"`
	CameraModel string `json:"cameraModel"`
	FNumber float64 `json:"fNumber"`
	ExposureDenominator float64 `json:"exposureDenominator"`
	ExposureNumerator float64 `json:"exposureNumerator"`
	FocalLength float64 `json:"focalLength"`
	Iso float64 `json:"iso"`
	TakenDateTime time.Time `json:"takenDateTime"` // timestamp
}

type Quota struct {
	Remaining float64 `json:"remaining"`
	Deleted float64 `json:"deleted"`
	State string `json:"state"` // normal | nearing | critical | exceeded
	Total float64 `json:"total"`
	Used float64 `json:"used"`
}

type SharingLink struct {
	Type string `json:"type"` // view | edit | embed | mail
	Token string `json:"token"`
	WebUrl string `json:"webUrl"`
	Application *Identity `json:"application"`
}

type Thumbnail struct {
	Width float64 `json:"width"`
	Height float64 `json:"height"`
	Url string `json:"url"` // url
}

type ThumbnailSet struct {
	Id string `json:"id"`
	Small *Thumbnail `json:"small"`
	Medium *Thumbnail `json:"medium"`
	Large *Thumbnail `json:"large"`
}

type UploadSession struct {
	ExpirationDateTime time.Time `json:"expirationDateTime"` // string timestamp
	NextExpectedRanges []string `json:"nextExpectedRanges"`
	UploadUrl string `json:"uploadUrl"`
}

type Video struct {
	Bitrate float64 `json:"bitrate"`
	Duration float64 `json:"duration"`
	Height float64 `json:"height"`
	Width float64 `json:"width"`
}

type ViewChanges struct {
	Instancechanges_resync string `json:"@changes.resync"`
	Value []*Item `json:"value"`
	Instanceodata_nextLink string `json:"@odata.nextLink"` // url
	Instancechanges_hasMoreChanges bool `json:"@changes.hasMoreChanges"`
	Instancechanges_token string `json:"@changes.token"`
}

