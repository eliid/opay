package order

import (
	"errors"
	"sync"
	"time"

	"github.com/henrylee2cn/opay"
	"github.com/jmoiron/sqlx"
)

// Order base model
type (
	BaseOrder struct {
		Id string `json:"id"`
		//second party's order id
		Id2 string `json:"id2,omitempty"`
		//asset id
		Aid string `json:"aid"`
		//second party's asset id
		Aid2 string `json:"aid2,omitempty"`
		//user id
		Uid string `json:"uid"`
		//second party's user id
		Uid2 string `json:"uid2,omitempty"`
		//order type
		Type uint8 `json:"type"`
		//the amount of change for the Uid-Aid account, balance of positive and negative representation
		Amount float64 `json:"amount"`
		//the amount of change for the Uid-Aid account, balance of positive and negative representation
		Amount2 float64 `json:"amount2,omitempty"`

		Summary   string    `json:"summary"`
		Details   []*Detail `json:"details"`
		Status    int32     `json:"status"`
		CreatedAt int64     `json:"created_at"`

		//processing error
		err error
		//database table name
		tableName string
		lock      sync.RWMutex
	}
	Detail struct {
		UpdatedAt int64  `json:"updated_at"`
		Status    int32  `json:"status"`
		Notes     string `json:"notes"`
		Ip        string `json:"ip"`
	}
)

var _ opay.IOrder = new(BaseOrder)

// 新建订单
func NewOrder(
	uid string,
	typ uint8,
	status int32,
	summary string,
	notes string,
	ip string,
) *BaseOrder {
	return (&BaseOrder{
		Uid:     uid,
		Type:    typ,
		Status:  status,
		Summary: summary,
	}).appendDetail(status, notes, ip)
}

// Get the most recent Action, the default value is UNSET==0.
func (this *BaseOrder) LastAction() opay.Action {
	return opay.Action(this.Status)
}

// Get user's id.
func (this *BaseOrder) GetUid() string {
	return this.Uid
}

// Get the second party's user id.
func (this *BaseOrder) GetUid2() string {
	return this.Uid2
}

// Get asset id.
func (this *BaseOrder) GetAid() string {
	return this.Aid
}

// Get the second party's asset id. (for example, the currency exchange business)
func (this *BaseOrder) GetAid2() string {
	return this.Aid2
}

// Get the amount of change for the Uid-Aid account,
// balance of positive and negative representation.
func (this *BaseOrder) GetAmount() float64 {
	return this.Amount
}

// Get the amount of change for the Uid-Aid2 account,
// balance of positive and negative representation.
func (this *BaseOrder) GetAmount2() float64 {
	return this.Amount2
}

// Async execution, and mark pending.
func (this *BaseOrder) ToPend(tx *sqlx.Tx, addition interface{}) error {
	return errors.New("This method 'ToPend' is not yet implemented.")
}

// Async execution, and mark the doing.
func (this *BaseOrder) ToDo(tx *sqlx.Tx, addition interface{}) error {
	return errors.New("This method 'ToDo' is not yet implemented.")
}

// Async execution, and mark the successful.
func (this *BaseOrder) ToSucceed(tx *sqlx.Tx, addition interface{}) error {
	return errors.New("This method 'ToSucceed' is not yet implemented.")
}

// Async execution, and mark canceled.
func (this *BaseOrder) ToCancel(tx *sqlx.Tx, addition interface{}) error {
	return errors.New("This method 'ToCancel' is not yet implemented.")
}

// Async execution, and mark failure.
func (this *BaseOrder) ToFail(tx *sqlx.Tx, addition interface{}) error {
	return errors.New("This method 'ToFail' is not yet implemented.")
}

// Sync execution, and mark the successful.
func (this *BaseOrder) SyncDeal(tx *sqlx.Tx, addition interface{}) error {
	return errors.New("This method 'SyncDeal' is not yet implemented.")
}

// Get error message.
func (this *BaseOrder) Err() error {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.err
}

// Writeback error message.
func (this *BaseOrder) SetErr(err error) {
	this.lock.Lock()
	this.err = err
	this.lock.Unlock()
}

// Get the order's id.
func (this *BaseOrder) GetId() string {
	return this.Id
}

// Get the second party's order id.
func (this *BaseOrder) GetId2() string {
	return this.Id2
}

// Get the order's type.
func (this *BaseOrder) GetType() uint8 {
	return this.Type
}

// Get status text.
func (this *BaseOrder) GetStatusText() string {
	return GetStatusText(this.Type, this.Status)
}

// Get the order's summary.
func (this *BaseOrder) GetSummary() string {
	return this.Summary
}

// Get the order processing record details.
func (this *BaseOrder) GetDetails() []*Detail {
	return this.Details
}

// Get the order's status.
func (this *BaseOrder) GetStatus() int32 {
	return this.Status
}

// Get the order's created time.
func (this *BaseOrder) GetCreatedAt() int64 {
	return this.CreatedAt
}

// Binding the order and it's related order.
func (this *BaseOrder) Bind(other *BaseOrder) {
	this.Id2, this.Uid2 = other.Id, other.Uid
	other.Id2, other.Uid2 = this.Id, this.Uid
}

// set order id, 32bytes(time23+type3+random6)
func (this *BaseOrder) setId() *BaseOrder {
	this.Id = CreateOrderid(this.Type)
	return this
}

// append order detail.
func (this *BaseOrder) appendDetail(status int32, notes string, ip string) *BaseOrder {
	this.Status = status
	if len(notes) == 0 {
		notes = this.GetStatusText()
	}
	this.Details = append(this.Details, &Detail{
		UpdatedAt: time.Now().Unix(),
		Status:    status,
		Notes:     notes,
		Ip:        ip,
	})
	return this
}