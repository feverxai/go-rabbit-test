package url

import "testing"

func Test_checkBlockList(t *testing.T) {
	type args struct {
		value interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"should return error",
			args{value: "https://www.facebook.com/"},
			true,
		},
		{
			"should not return error",
			args{value: "https://www.google.com/"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkBlockList(tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("checkBlockList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_generateRandomString(t *testing.T) {
	type args struct {
		length int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"should contains 8 random character",
			args{length: 8},
			"12345678",
		},
		{
			"should contains 4 random character",
			args{length: 4},
			"1234",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateRandomString(tt.args.length); len(got) != len(tt.want) {
				t.Errorf("generateRandomString() = %v, want %v", got, tt.want)
			}
		})
	}
}
