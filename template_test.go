package sdk

import "testing"

func TestTpl(t *testing.T) {
	type args struct {
		url  string
		vars map[string]string
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: "main",
		args: args{
			url:  "/tenants/{tenant}/apps/{app}",
			vars: map[string]string{"app": "app_id", "tenant": "tenant_id"},
		},
		want: "/tenants/tenant_id/apps/app_id",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Tpl(tt.args.url, tt.args.vars); got != tt.want {
				t.Errorf("Tpl() = %v, want %v", got, tt.want)
			}
		})
	}
}
