package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPostgresUtils(t *testing.T) {
	id := uuid.New()
	pgUUID := MapToPgUUID(id)
	if !pgUUID.Valid || uuid.UUID(pgUUID.Bytes) != id {
		t.Errorf("UUID mapping failed")
	}

	pgText := MapToPgText("sample")
	if !pgText.Valid || pgText.String != "sample" {
		t.Errorf("MapToPgText failed")
	}

	nilPtrText := MapPtrToPgText(nil)
	if nilPtrText.Valid {
		t.Errorf("expected nilPtrText to be invalid")
	}

	strVal := "hello"
	ptrText := MapPtrToPgText(&strVal)
	if !ptrText.Valid || ptrText.String != "hello" {
		t.Errorf("MapPtrToPgText failed")
	}

	now := time.Now().Truncate(time.Microsecond)
	pgTime := MapToPgTimestamptz(now)
	if !pgTime.Valid || !pgTime.Time.Equal(now) {
		t.Errorf("MapToPgTimestamptz failed")
	}

	val := 123.45
	numeric, err := MapFloat64ToPgNumeric(val, nil)
	if err != nil {
		t.Fatalf("MapFloat64ToPgNumeric failed: %v", err)
	}
	floatVal, err := MapPgNumericToFloat64(numeric)
	if err != nil {
		t.Fatalf("MapPgNumericToFloat64 failed: %v", err)
	}
	if floatVal != val {
		t.Errorf("expected %f, got %f", val, floatVal)
	}

	intVal := int64(98765)
	intNumeric := MapInt64ToPgNumeric(intVal, 0)
	gotInt, err := MapPgNumericToInt64(intNumeric, 0)
	if err != nil {
		t.Fatalf("MapPgNumericToInt64 failed: %v", err)
	}
	if gotInt != intVal {
		t.Errorf("expected %d, got %d", intVal, gotInt)
	}
}
