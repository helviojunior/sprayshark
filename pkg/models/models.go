package models

import (
	"time"
	"fmt"
)

// RequestType are network log types
type RequestType int

const (
	HTTP RequestType = 0
	WS
)
// Result is a github.com/helviojunior/spraysharksprayshark result
type Result struct {
	ID uint `json:"id" gorm:"primarykey"`

	TestId                string    `json:"test_id"`
	User                  string    `json:"username"`
	Password              string    `json:"password"`
	PasswordHash          string    `json:"password_hash"`
	ProbedAt              time.Time `json:"probed_at"`

	UserExists       	  bool   	`json:"user_exists"`
	ValidCredential    	  bool   	`json:"valid_credential"`

	Screenshot            string    `json:"screenshot"`
	HTML                  string    `json:"html" gorm:"index"`

	// Name of the screenshot file
	Filename string `json:"file_name"`

	// Failed flag set if the result should be considered failed
	Failed       bool   `json:"failed"`
	FailedReason string `json:"failed_reason"`

	Network []NetworkLog `json:"network" gorm:"constraint:OnDelete:CASCADE"`

}

func (result *Result) CalcHash() string {
	var hash uint64
	var mask uint64
	var bits uint64
	hash = 0
	bits = 13
	mask = 0xFFFFFFFF
	for _, b := range result.Password {
		c := uint64(b)
		hash += (c >> bits | c << (32 - bits)) & mask;
	}
	hash = hash & mask;

	result.PasswordHash = fmt.Sprintf("%06d", hash)
	return result.PasswordHash
}

type NetworkLog struct {
	ID       uint `json:"id" gorm:"primarykey"`
	ResultID uint `json:"result_id"`

	RequestType RequestType `json:"request_type"`
	StatusCode  int64       `json:"status_code"`
	URL         string      `json:"url"`
	RemoteIP    string      `json:"remote_ip"`
	MIMEType    string      `json:"mime_type"`
	Time        time.Time   `json:"time"`
	Content     []byte      `json:"content"`
	Error       string      `json:"error"`
}