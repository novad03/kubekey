package modules

import (
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"testing"
)

func TestTask_calculateConcurrency(t1 *testing.T) {
	type fields struct {
		Hosts       []connector.BaseHost
		Concurrency float64
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "test1",
			fields: fields{
				Concurrency: 0.5,
				Hosts: []connector.BaseHost{
					{Name: "node1"},
					{Name: "node2"},
					{Name: "node3"},
				},
			},
			want: 2,
		},
		{
			name: "test2",
			fields: fields{
				Concurrency: 0.5,
				Hosts: []connector.BaseHost{
					{Name: "node1"},
					{Name: "node2"},
					{Name: "node3"},
					{Name: "node4"},
				},
			},
			want: 2,
		},
		{
			name: "test3",
			fields: fields{
				Concurrency: 0.4,
				Hosts: []connector.BaseHost{
					{Name: "node1"},
					{Name: "node2"},
					{Name: "node3"},
				},
			},
			want: 1,
		},
		{
			name: "test4",
			fields: fields{
				Concurrency: 0.4,
				Hosts: []connector.BaseHost{
					{Name: "node1"},
					{Name: "node2"},
					{Name: "node3"},
					{Name: "node4"},
				},
			},
			want: 2,
		},
		{
			name: "test5",
			fields: fields{
				Concurrency: 0.1,
				Hosts: []connector.BaseHost{
					{Name: "node1"},
					{Name: "node2"},
					{Name: "node3"},
					{Name: "node4"},
				},
			},
			want: 1,
		},
		{
			name: "test6",
			fields: fields{
				Concurrency: 0.222222222222222,
				Hosts: []connector.BaseHost{
					{Name: "node1"},
					{Name: "node2"},
					{Name: "node3"},
					{Name: "node4"},
				},
			},
			want: 1,
		},
		{
			name: "test7",
			fields: fields{
				Concurrency: 1,
				Hosts: []connector.BaseHost{
					{Name: "node1"},
					{Name: "node2"},
					{Name: "node3"},
					{Name: "node4"},
				},
			},
			want: 4,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Task{
				Concurrency: tt.fields.Concurrency,
				//Hosts:       []runtimes.Host(tt.fields.Hosts),
			}
			if got := t.calculateConcurrency(); got != tt.want {
				t1.Errorf("calculateConcurrency() = %v, want %v", got, tt.want)
			}
		})
	}
}
