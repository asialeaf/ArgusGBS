<template>
  <div
    class="el-radio-group"
    role="radiogroup"
    @keydown="handleKeydown"
  >
    <slot></slot>
  </div>
</template>
<script>
  import Emitter from 'element-ui/src/mixins/emitter';

  const keyCode = Object.freeze({
    LEFT: 37,
    UP: 38,
    RIGHT: 39,
    DOWN: 40
  });
  export default {
    name: 'ElRadioGroup',

    componentName: 'ElRadioGroup',

    inject: {
      elFormItem: {
        default: ''
      }
    },

    mixins: [Emitter],

    props: {
      value: {},
      size: String,
      fill: String,
      textColor: String,
      disabled: Boolean
    },

    computed: {
      _elFormItemSize() {
        return (this.elFormItem || {}).elFormItemSize;
      },
      radioGroupSize() {
        return this.size || this._elFormItemSize || (this.$ELEMENT || {}).size;
      }
    },

    created() {
      this.$on('handleChange', value => {
        this.$emit('change', value);
      });
    },
    mounted() {
      // 当radioGroup没有默认选项时，第一个可以选中Tab导航
      const radios = this.$el.querySelectorAll('[type=radio]');
      if (![].some.call(radios, radio => radio.checked)) {
        const first = [].find.call(radios, radio => !radio.disabled);
        if (first) first.tabIndex = 0;
      }
    },
    methods: {
      handleKeydown(e) { // 左右上下按键 可以在radio组内切换不同选项
        const step = {
          [keyCode.LEFT]: -1,
          [keyCode.UP]: -1,
          [keyCode.RIGHT]: 1,
          [keyCode.DOWN]: 1
        }[e.keyCode];
        if (step === undefined) return;
        e.stopPropagation();
        e.preventDefault();
        const radios = this.$el.querySelectorAll('[type=radio]');
        const length = radios.length;
        if (length === 0) return;
        // 焦点可能在原生radio上，也可能在radio-button的label上（label包裹着input）
        const current = e.target.nodeName === 'INPUT'
          ? e.target
          : (e.target.querySelector && e.target.querySelector('input[type=radio]'));
        const index = [].indexOf.call(radios, current);
        if (index < 0) return;
        // 沿方向查找第一个可用的radio，跨过disabled项
        for (let i = 1; i <= length; i++) {
          const target = radios[(index + step * i + length) % length];
          if (!target.disabled) {
            target.click();
            target.focus();
            break;
          }
        }
      }
    },
    watch: {
      value(value) {
        this.dispatch('ElFormItem', 'el.form.change', [this.value]);
      }
    }
  };
</script>

