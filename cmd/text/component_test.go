package text

import "testing"

func TestApt_Docker(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "version",
			args: args{version: "18.09"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap := Apt{}
			got, err := ap.Docker(tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("Docker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Docker() got = %v, want %v", got, tt.want)
			}
		})
	}
}
