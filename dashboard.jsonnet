local grafana = import 'grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local template = grafana.template;
local singlestat = grafana.singlestat;
local graphPanel = grafana.graphPanel;
local prometheus = grafana.prometheus;
local text = grafana.text;

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
  'nvidia_clocks_throttle_reasons_gpu_idle',
  'nvidia_clocks_throttle_reasons_hw_power_brake_slowdown',
  'nvidia_clocks_throttle_reasons_hw_slowdown',
  'nvidia_clocks_throttle_reasons_hw_thermal_slowdown',
  'nvidia_clocks_throttle_reasons_sw_thermal_slowdown',
  'nvidia_clocks_throttle_reasons_sync_boost',
  'nvidia_clocks_video',
  'nvidia_driver_version',
  'nvidia_ecc_errors_corrected_aggregate_device_memory',
  'nvidia_ecc_errors_corrected_aggregate_l1_cache',
  'nvidia_ecc_errors_corrected_aggregate_l2_cache',
  'nvidia_ecc_errors_corrected_aggregate_register_file',
  'nvidia_ecc_errors_corrected_aggregate_texture_memory',
  'nvidia_ecc_errors_corrected_aggregate_total',
  'nvidia_ecc_errors_corrected_volatile_l1_cache',
  'nvidia_ecc_errors_corrected_volatile_l2_cache',
  'nvidia_ecc_errors_corrected_volatile_register_file',
  'nvidia_ecc_errors_corrected_volatile_texture_memory',
  'nvidia_ecc_errors_corrected_volatile_total',
  'nvidia_ecc_errors_uncorrected_aggregate_device_memory',
  'nvidia_ecc_errors_uncorrected_aggregate_l1_cache',
  'nvidia_ecc_errors_uncorrected_aggregate_l2_cache',
  'nvidia_ecc_errors_uncorrected_aggregate_register_file',
  'nvidia_ecc_errors_uncorrected_aggregate_texture_memory',
  'nvidia_ecc_errors_uncorrected_aggregate_total',
  'nvidia_ecc_errors_uncorrected_volatile_device_memory',
  'nvidia_ecc_errors_uncorrected_volatile_l1_cache',
  'nvidia_ecc_errors_uncorrected_volatile_l2_cache',
  'nvidia_ecc_errors_uncorrected_volatile_register_file',
  'nvidia_ecc_errors_uncorrected_volatile_texture_memory',
  'nvidia_ecc_errors_uncorrected_volatile_total',
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
  'nvidia_pstate_unparseable',
  'nvidia_query_field_unsupported',
  'nvidia_retired_pages_dbe',
  'nvidia_retired_pages_double_bit_count',
  'nvidia_retired_pages_pending',
  'nvidia_retired_pages_sbe',
  'nvidia_retired_pages_single_bit_ecc_count',
  'nvidia_temperature_gpu',
  'nvidia_unknown_error',
  'nvidia_unparseable_query_result_value',
  'nvidia_utilization_gpu',
  'nvidia_utilization_memory',
];

local dashboardWitdh = 24;
local panelSize = {
  height: 8,
  width: 8,
};

local notePanel = text.new(
  span=5,
  mode='markdown',
  content='## nvidia-smi-exporter\nSome panels within this dashboard may not be populated if the GPU being scraped do not support those metrics.',
  transparent=false
);

local graphPanels = std.mapWithIndex(function(i, metric)
  graphPanel.new(
    title=metric,
    datasource='Prometheus',
    linewidth=1,
  ).addTarget(
    prometheus.target(
      metric,
      intervalFactor=1,  // resolution factor where 2 => 1/2
      interval='1s',  // minStep
    )
  ), metrics);

dashboard.new(
  'NVIDIA GPU',
  tags=['nvidia'],
  description='Dashboard for the nvidia-smi-exporter. Some panels may not contain data for unsupported nvidia-smi query fields.',
  schemaVersion=18,
  editable=false,
  time_from='now-30m',
  refresh='30s',
  graphTooltip='shared_crosshair',
  uid='gpu',
)
.addPanels(
  std.mapWithIndex(function(i, panel)
    panel { gridPos: {
      h: panelSize.height,
      w: panelSize.width,
      x: i * panelSize.width % dashboardWitdh,
      y: panelSize.height * std.floor(i * panelSize.width / dashboardWitdh),
    } }, [notePanel] + graphPanels)
)
