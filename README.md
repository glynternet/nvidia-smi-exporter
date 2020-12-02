# nvidia-smi-exporter

Cross-platform nvidia-smi metrics prometheus exporter and grafana dashboard

## Build
```
# Set OS to change operating system, default linux
# Set ARCH to change architecture, default amd64
make nvidia-smi-exporter-binary
# builds to -o build/VERSION/OS-ARCH/nvidia-smi-exporter
```

## Run
```
# Ensure nvidia-smi is in your PATH or exporter NVIDIA_SMI with path to the binary
nvidia-smi-exporter serve
```
Default port is 9101

## Launch at startup on Windows
1. Build exporter for Windows
2. Move exporter to `C:\\Windows\System32`
3. Create batch file in `C:\\Windows\System32` containing `nvidia-smi-exporter.exe serve`
4. Hit `WINDOWS_KEY + R` then run `shell:common startup`. This should open windows explorer.
5. Drag the batch file into that explorer window. This should create a shortcut, not move the original.
6. Restart and check that metrics are scrapable.

### Metrics Reported per GPU
See [Dashboard section](#dashboard) to find out about removing unsupported metrics from your dashboard.

```
nvidia_clocks_applications_gr
nvidia_clocks_applications_graphics
nvidia_clocks_applications_mem
nvidia_clocks_applications_memory
nvidia_clocks_current_graphics
nvidia_clocks_current_memory
nvidia_clocks_current_sm
nvidia_clocks_current_video
nvidia_clocks_default_applications_gr
nvidia_clocks_default_applications_graphics
nvidia_clocks_default_applications_mem
nvidia_clocks_default_applications_memory
nvidia_clocks_gr
nvidia_clocks_max_gr
nvidia_clocks_max_graphics
nvidia_clocks_max_mem
nvidia_clocks_max_memory
nvidia_clocks_max_sm
nvidia_clocks_mem
nvidia_clocks_sm
nvidia_clocks_throttle_reasons_gpu_idle
nvidia_clocks_throttle_reasons_hw_power_brake_slowdown
nvidia_clocks_throttle_reasons_hw_slowdown
nvidia_clocks_throttle_reasons_hw_thermal_slowdown
nvidia_clocks_throttle_reasons_sw_thermal_slowdown
nvidia_clocks_throttle_reasons_sync_boost
nvidia_clocks_video
nvidia_driver_version
nvidia_ecc_errors_corrected_aggregate_device_memory
nvidia_ecc_errors_corrected_aggregate_l1_cache
nvidia_ecc_errors_corrected_aggregate_l2_cache
nvidia_ecc_errors_corrected_aggregate_register_file
nvidia_ecc_errors_corrected_aggregate_texture_memory
nvidia_ecc_errors_corrected_aggregate_total
nvidia_ecc_errors_corrected_volatile_l1_cache
nvidia_ecc_errors_corrected_volatile_l2_cache
nvidia_ecc_errors_corrected_volatile_register_file
nvidia_ecc_errors_corrected_volatile_texture_memory
nvidia_ecc_errors_corrected_volatile_total
nvidia_ecc_errors_uncorrected_aggregate_device_memory
nvidia_ecc_errors_uncorrected_aggregate_l1_cache
nvidia_ecc_errors_uncorrected_aggregate_l2_cache
nvidia_ecc_errors_uncorrected_aggregate_register_file
nvidia_ecc_errors_uncorrected_aggregate_texture_memory
nvidia_ecc_errors_uncorrected_aggregate_total
nvidia_ecc_errors_uncorrected_volatile_device_memory
nvidia_ecc_errors_uncorrected_volatile_l1_cache
nvidia_ecc_errors_uncorrected_volatile_l2_cache
nvidia_ecc_errors_uncorrected_volatile_register_file
nvidia_ecc_errors_uncorrected_volatile_texture_memory
nvidia_ecc_errors_uncorrected_volatile_total
nvidia_encoder_stats_averageFps
nvidia_encoder_stats_averageLatency
nvidia_encoder_stats_sessionCount
nvidia_enforced_power_limit
nvidia_fan_speed
nvidia_memory_free
nvidia_memory_total
nvidia_memory_used
nvidia_power_default_limit
nvidia_power_draw
nvidia_power_limit
nvidia_power_management
nvidia_power_max_limit
nvidia_power_min_limit
nvidia_pstate
nvidia_pstate_unparseable
nvidia_query_field_unsupported
nvidia_retired_pages_dbe
nvidia_retired_pages_double_bit_count
nvidia_retired_pages_pending
nvidia_retired_pages_sbe
nvidia_retired_pages_single_bit_ecc_count
nvidia_temperature_gpu
nvidia_unknown_error
nvidia_unparseable_query_result_value
nvidia_utilization_gpu
nvidia_utilization_memory
```

### Dashboard
The dashboard is generated using jsonnet and the grafana/grafonnet-lib library.
If your GPU does not support all of the metrics, you may want to edit the `dashboard.jsonnet` file to remove and/or reorder some of the metric names. Then generate the dashboard using the following command:

```shell
jsonnet -J ../../grafana/grafonnet-lib ./dashboard.jsonnet > ./dashboard.json
```

### Prometheus example config

```yaml
- job_name: "nvidia_gpu"
  static_configs:
  - targets: ['HOST:9101'] # default port is 9101
```
