local grafana = import 'grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local template = grafana.template;
local singlestat = grafana.singlestat;
local graphPanel = grafana.graphPanel;
local prometheus = grafana.prometheus;

local metrics = [
  'nvidia_clocks_applications_gr',
  'nvidia_clocks_applications_graphics',
  'nvidia_clocks_applications_mem',
  'nvidia_clocks_applications_memory',
  'nvidia_clocks_current_graphics',
  'nvidia_clocks_current_memory',
  'nvidia_clocks_current_sm',
  'nvidia_clocks_current_video',
  'nvidia_clocks_default_applications_gr',
  'nvidia_clocks_default_applications_graphics',
  'nvidia_clocks_default_applications_mem',
  'nvidia_clocks_default_applications_memory',
  'nvidia_clocks_gr',
  'nvidia_clocks_max_gr',
  'nvidia_clocks_max_graphics',
  'nvidia_clocks_max_mem',
  'nvidia_clocks_max_memory',
  'nvidia_clocks_max_sm',
  'nvidia_clocks_mem',
  'nvidia_clocks_sm',
  'nvidia_clocks_video',
  'nvidia_driver_version',
  'nvidia_encoder_stats_averageFps',
  'nvidia_encoder_stats_averageLatency',
  'nvidia_encoder_stats_sessionCount',
  'nvidia_enforced_power_limit',
  'nvidia_fan_speed',
  'nvidia_memory_free',
  'nvidia_memory_total',
  'nvidia_memory_used',
  'nvidia_power_default_limit',
  'nvidia_power_draw',
  'nvidia_power_limit',
  'nvidia_power_management',
  'nvidia_power_max_limit',
  'nvidia_power_min_limit',
  'nvidia_pstate',
  'nvidia_temperature_gpu',
  'nvidia_utilization_gpu',
  'nvidia_utilization_memory',
  'nvidia_unknown_error',
  'nvidia_query_field_unsupported',
];

local dashboardWitdh = 24;
local panelSize = {
  height: 8,
  width: 8,
};

dashboard.new(
  'NVIDIA GPU',
  tags=['nvidia'],
  schemaVersion=18,
  editable=false,
  time_from='now-30m',
  refresh='15s',
  graphTooltip='shared_crosshair',
  uid='gpu',
)
.addPanels(
  [graphPanel.new(
    title=metrics[metricIndex],
    datasource='Prometheus',
    linewidth=2,
  ).addTarget(
    prometheus.target(metrics[metricIndex])
  ) { gridPos: {
    h: panelSize.height,
    w: panelSize.width,
    x: metricIndex * panelSize.width % dashboardWitdh,
    y: panelSize.height * std.floor(metricIndex * panelSize.width / dashboardWitdh),
  } } for metricIndex in std.range(0, std.length(metrics) - 1)]
)
