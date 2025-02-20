<script lang="ts">
import { Banner } from '@rancher/components';
import isEmpty from 'lodash/isEmpty';
import Vue from 'vue';
import { mapGetters } from 'vuex';

import EmptyState from '@pkg/components/EmptyState.vue';
import SnapshotCard from '@pkg/components/SnapshotCard.vue';
import { Snapshot, SnapshotEvent } from '@pkg/main/snapshots/types';
import { ipcRenderer } from '@pkg/utils/ipcRenderer';

interface Data {
  snapshotEvent: SnapshotEvent | null;
  snapshotsPollingInterval: ReturnType<typeof setInterval> | null;
  isEmpty: boolean;
}

interface Methods {
  pollingStart: () => void,
}

interface Computed {
  snapshots: Snapshot[],
}

export default Vue.extend<Data, Methods, Computed, never>({
  components: {
    Banner,
    EmptyState,
    SnapshotCard,
  },

  data(): Data {
    return {
      snapshotsPollingInterval: null,
      snapshotEvent:            null,
      isEmpty:                  false,
    };
  },

  computed: { ...mapGetters('snapshots', { snapshots: 'list' }) },

  watch: {
    snapshots(list) {
      this.isEmpty = list?.length === 0;
    },
  },

  beforeMount() {
    this.$store.dispatch('snapshots/fetch');
    this.pollingStart();

    ipcRenderer.on('snapshot', (_, event) => {
      this.snapshotEvent = event;
    });

    if (isEmpty(this.$route.params)) {
      return;
    }

    const { type, result, snapshotName } = this.$route.params as SnapshotEvent;

    this.snapshotEvent = {
      type, result, snapshotName,
    };
  },

  beforeDestroy() {
    if (this.snapshotsPollingInterval) {
      clearInterval(this.snapshotsPollingInterval);
    }
    ipcRenderer.removeAllListeners('snapshot');
  },

  methods: {
    pollingStart() {
      this.snapshotsPollingInterval = setInterval(() => {
        this.$store.dispatch('snapshots/fetch');
      }, 1500);
    },
  },
});
</script>

<template>
  <div class="snapshots">
    <Banner
      v-if="snapshotEvent"
      class="banner mb-20"
      :color="snapshotEvent.result"
      :closable="true"
      @close="snapshotEvent=null"
    >
      <span
        class="event-message"
        v-html="t(`snapshots.info.${ snapshotEvent.type }.${ snapshotEvent.result }`, { snapshot: snapshotEvent.snapshotName, error: snapshotEvent.error }, true)"
      />
    </Banner>
    <div
      v-for="(item) of snapshots"
      :key="item.name"
    >
      <SnapshotCard
        class="mb-20"
        :value="item"
      />
    </div>
    <div v-if="isEmpty">
      <empty-state
        class="mt-10"
        :icon="t('snapshots.empty.icon')"
        :heading="t('snapshots.empty.heading')"
        :body="t('snapshots.empty.body')"
      >
      </empty-state>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .event-message {
    word-wrap: break-word;
    overflow: hidden;
    text-overflow: ellipsis;
  }
</style>
