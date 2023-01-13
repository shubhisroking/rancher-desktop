<script lang="ts">
import { Checkbox } from '@rancher/components';
import Vue from 'vue';

import RdFieldset from '@pkg/components/form/RdFieldset.vue';
import { Settings } from '@pkg/config/settings';
import { RecursiveTypes } from '@pkg/utils/typeUtils';

import type { PropType } from 'vue';

export default Vue.extend({
  name:       'preferences-application-behavior',
  components: { Checkbox, RdFieldset },
  props:      {
    preferences: {
      type:     Object as PropType<Settings>,
      required: true,
    },
  },
  methods: {
    onChange<P extends keyof RecursiveTypes<Settings>>(property: P, value: RecursiveTypes<Settings>[P]) {
      this.$store.dispatch('preferences/updatePreferencesData', { property, value });
    },
    onAdminAccessChange(val: boolean) {
      this.$store.dispatch('applicationSettings/commitAdminAccess', val);
    },
  },
});
</script>

<template>
  <div class="application-behavior">
    <rd-fieldset
      data-test="autoStart"
      :legend-text="t('application.behavior.autoStart.legendText')"
    >
      <checkbox
        :label="t('application.behavior.autoStart.label')"
        :value="preferences.autoStart"
        @input="onChange('autoStart', $event)"
      />
    </rd-fieldset>
    <rd-fieldset
      data-test="autoStart"
      :legend-text="t('application.behavior.autoStart.legendText')"
    >
      <checkbox
        label="Allow Rancher Desktop to acquire administrative credentials (sudo access)"
        :value="hasAdminAccess"
        @input="onChange('application.adminAccess', $event)"
      />
    </rd-fieldset>
    <rd-fieldset
      data-test="background"
      :legend-text="t('application.behavior.background.legendText')"
      :legend-tooltip="t('application.behavior.background.legendTooltip')"
      class="checkbox-group"
    >
      <checkbox
        data-test="automaticUpdatesCheckbox"
        label="Check for updates automatically"
        :value="canAutoUpdate"
        @input="onChange('application.updater.enabled', $event)"
      />
      <checkbox
        :label="t('application.behavior.startInBackground.label')"
        :value="preferences.startInBackground"
        @input="onChange('startInBackground', $event)"
      />
      <checkbox
        :label="t('application.behavior.windowQuitOnClose.label')"
        :value="preferences.window.quitOnClose"
        @input="onChange('window.quitOnClose', $event)"
      />
    </rd-fieldset>
    <rd-fieldset
      data-test="notificationIcon"
      :legend-text="t('application.behavior.notificationIcon.legendText')"
    >
      <checkbox
        label="Allow collection of anonymous statistics to help us improve Rancher Desktop"
        :value="preferences.telemetry"
        @input="onChange('telemetry', $event)"
      />
    </rd-fieldset>
    <rd-fieldset
      data-test="notificationIcon"
      :legend-text="t('application.behavior.notificationIcon.legendText')"
    >
      <checkbox
        :label="t('application.behavior.notificationIcon.label')"
        :value="preferences.hideNotificationIcon"
        @input="onChange('hideNotificationIcon', $event)"
      />
    </rd-fieldset>
    <rd-fieldset
      data-test="statistics"
      legend-text="Statistics"
    >
      <checkbox
        label="Allow collection of anonymous statistics to help us improve Rancher Desktop"
        :value="preferences.application.telemetry.enabled"
        @input="onChange('application.telemetry.enabled', $event)"
      />
    </rd-fieldset>
  </div>
</template>

<style lang="scss" scoped>
  .application-behavior {
    display: flex;
    flex-direction: column;
    gap: 1rem;

    .checkbox-group {
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
    }
  }
</style>
