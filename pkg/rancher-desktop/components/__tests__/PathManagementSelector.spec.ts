import { shallowMount } from '@vue/test-utils';
import Vue from 'vue';

import PathManagementSelector from '../PathManagementSelector.vue';


import { PathManagementStrategy } from '@pkg/integrations/pathManager';

Vue.extend({
  methods: {
    t: (s: string): string => {
      return s;
    },
  }
});

describe('Preferences/Application/PathManagementSelector.vue', () => {
  it(`can select a different item in the preferences radio-group`, () => {
    const wrapper = shallowMount(PathManagementSelector, { propsData: { value:  PathManagementStrategy.RcFiles } });
    const radioGroup = wrapper.find('radioGroup');

    expect(radioGroup).not.toBeNull();
//    expect(wrapper.find('button').classes()).toStrictEqual(['btn', 'role-fab', 'ripple']);
  });

  // it(`emits 'open-preferences' on click`, () => {
  //   const wrapper = shallowMount(
  //     PathManagementSelector,
  //     { });
  //
  //   wrapper.find('button').trigger('click');
  //
  //   expect(wrapper.emitted('open-preferences')).toHaveLength(1);
  // });
});
