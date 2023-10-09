package cliniko

const (
	AttendeeCancelCancellationReasonFeelingBetter  = 10
	AttendeeCancelCancellationReasonConditionWorse = 20
	AttendeeCancelCancellationReasonSick           = 30
	AttendeeCancelCancellationReasonCOVID19Related = 31
	AttendeeCancelCancellationReasonAway           = 40
	AttendeeCancelCancellationReasonOther          = 50
	AttendeeCancelCancellationReasonWork           = 60
)

type AttendeeCancelCancellationReason int

const (
	ContactPhoneNumbersPhoneTypeMobile = "Mobile"
	ContactPhoneNumbersPhoneTypeHome   = "Home"
	ContactPhoneNumbersPhoneTypeWork   = "Work"
	ContactPhoneNumbersPhoneTypeOther  = "Other"
	ContactPhoneNumbersPhoneTypeFax    = "Fax"
)

type ContactPhoneNumbersPhoneType string

const (
	IndividualAppointmentCancelCancellationReasonFeelingBetter  = 10
	IndividualAppointmentCancelCancellationReasonConditionWorse = 20
	IndividualAppointmentCancelCancellationReasonSick           = 30
	IndividualAppointmentCancelCancellationReasonCOVID19Related = 31
	IndividualAppointmentCancelCancellationReasonAway           = 40
	IndividualAppointmentCancelCancellationReasonOther          = 50
	IndividualAppointmentCancelCancellationReasonWork           = 60
)

type IndividualAppointmentCancelCancellationReason int

const (
	PatientPatientPhoneNumbersPhoneTypeMobile = "Mobile"
	PatientPatientPhoneNumbersPhoneTypeHome   = "Home"
	PatientPatientPhoneNumbersPhoneTypeWork   = "Work"
	PatientPatientPhoneNumbersPhoneTypeOther  = "Other"
	PatientPatientPhoneNumbersPhoneTypeFax    = "Fax"
)

type PatientPatientPhoneNumbersPhoneType string

const (
	StockAdjustmentAdjustmentTypeStockPurchase = "Stock Purchase"
	StockAdjustmentAdjustmentTypeReturned      = "Returned"
	StockAdjustmentAdjustmentTypeOther         = "Other"
	StockAdjustmentAdjustmentTypeDamaged       = "Damaged"
	StockAdjustmentAdjustmentTypeOutOfDate     = "Out of Date"
	StockAdjustmentAdjustmentTypeItemSold      = "Item Sold"
)

type StockAdjustmentAdjustmentType string
