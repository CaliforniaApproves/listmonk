<template>
  <div class="modal-card content" style="width: auto">
    <header class="modal-card-head">
      <h4>{{ $t('lists.leaderboard') }}: {{ list.name }}</h4>
    </header>

    <section class="modal-card-body">
      <b-loading :active="isLoading" :is-full-page="false" />

      <b-table :data="entries" :narrowed="true" :striped="true" data-cy="leaderboard">
        <b-table-column v-slot="props" field="rank" label="#" width="8%">
          {{ props.row.rank }}
        </b-table-column>

        <b-table-column v-slot="props" field="name" :label="$t('globals.fields.name')">
          <router-link :to="`/subscribers/${props.row.subscriberId}`">
            {{ props.row.name }}
          </router-link>
          <br />
          <span class="is-size-7 has-text-grey">{{ props.row.email }}</span>
        </b-table-column>

        <b-table-column v-slot="props" field="referralCode" :label="$t('subscribers.referralCode')">
          <copy-text :text="props.row.referralCode" />
        </b-table-column>

        <b-table-column v-slot="props" field="count" :label="$t('lists.referrals')" numeric>
          <strong>{{ props.row.count }}</strong>
        </b-table-column>

        <template #empty v-if="!isLoading">
          <div class="has-text-centered has-text-grey">
            {{ $t('lists.noReferrals') }}
          </div>
        </template>
      </b-table>
    </section>

    <footer class="modal-card-foot has-text-right">
      <b-button @click="$parent.close()">
        {{ $t('globals.buttons.close') }}
      </b-button>
    </footer>
  </div>
</template>

<script>
import Vue from 'vue';
import CopyText from '../components/CopyText.vue';

export default Vue.extend({
  name: 'ListLeaderboard',

  components: { CopyText },

  props: {
    list: { type: Object, default: () => ({}) },
  },

  data() {
    return {
      entries: [],
      isLoading: false,
    };
  },

  mounted() {
    this.isLoading = true;
    this.$api.getListLeaderboard(this.list.id, { per_page: 50 }).then((data) => {
      this.entries = data;
    }).finally(() => {
      this.isLoading = false;
    });
  },
});
</script>
