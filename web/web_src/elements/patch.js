import Vue from "vue";
import ElCheckbox from "elements/checkbox.vue";
import ElRadio from "elements/radio.vue";
import ElRadioGroup from "elements/radio-group.vue";

export default function patchElementUI() {
    Vue.component("ElCheckbox", ElCheckbox);
    Vue.component("ElRadio", ElRadio);
    Vue.component("ElRadioGroup", ElRadioGroup);
}
