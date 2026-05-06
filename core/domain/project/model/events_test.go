package model_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/usegavel/gavel/core/domain/project/model"
)

func TestArchitecturePolicyUpdatedEvent(t *testing.T) {
	id := model.NewProjectID(uuid.New())
	evt := model.NewArchitecturePolicyUpdated(id, testTime)

	assert.Equal(t, model.EventNameArchitecturePolicyUpdated, evt.EventName())
	assert.Equal(t, testTime, evt.OccurredAt())
	assert.True(t, id.Equal(evt.ProjectID()))
}
