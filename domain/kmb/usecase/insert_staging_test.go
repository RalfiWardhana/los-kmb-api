package usecase

import (
	"errors"
	"los-kmb-api/domain/kmb/interfaces/mocks"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInsertStaging(t *testing.T) {
	testcases := []struct {
		name             string
		prospectID       string
		errSaveToStaging error
		result           response.InsertStaging
		err              error
	}{
		{
			name:       "test insert staging",
			prospectID: "TEST1",
			result: response.InsertStaging{
				ProspectID: "TEST1",
			},
		},
		{
			name:             "test insert staging err",
			errSaveToStaging: errors.New("insert staging err"),
			err:              errors.New(constant.ERROR_UPSTREAM + " - insert staging err"),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("SaveToStaging", tc.prospectID).Return(tc.errSaveToStaging)
			usecase := NewUsecase(mockRepository, mockHttpClient)

			result, err := usecase.InsertStaging(tc.prospectID)
			require.Equal(t, tc.result, result)
			require.Equal(t, tc.err, err)
		})
	}

}
