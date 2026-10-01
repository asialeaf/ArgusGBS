<template>
    <FormDlg title="签发国密证书" ref="dlg" @hide="onHide" @show="onShow" @submit="onSubmit" :disabled="errors.any()">
        <div :class="{'form-group': true, 'has-feedback': true,'has-error': errors.has('common_name')}" v-show="common_name">
            <label class="col-sm-4 control-label">签发给</label>
            <div class="col-sm-7">
                <input type="text" class="form-control" name="common_name" :value="common_name" readonly>
            </div>
        </div>
        <div :class="{'form-group': true, 'has-feedback': true,'has-error': errors.has('valid_days')}">
            <label class="col-sm-4 control-label">有效期(天)
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" name="valid_days"
                    v-model.trim="form.valid_days" data-vv-as="有效期" v-validate="'required|numeric|min_value:1'" @keydown.enter.prevent="onSubmit">
            </div>
        </div>
    </FormDlg>
</template>

<script>
import FormDlg from 'components/FormDlg.vue';
import $ from 'jquery';

export default {
    data() {
        return {
            common_name: "",
            form: this.defForm(),
        }
    },
    components: {
        FormDlg
    },
    methods: {
        defForm() {
            return {
                id: "",
                valid_days: 3650,
            }
        },
        onHide() {
            this.common_name = "";
            this.form = this.defForm();
        },
        onShow() {
            this.errors.clear();
        },
        async onSubmit() {
            var ok = await this.$validator.validateAll();
            if(!ok) {
                var e = this.errors.items[0];
                this.$message({
                    type: 'error',
                    message: e.msg
                })
                $(`[name=${e.field}]`).focus();
                return;
            }
            window.open(`/api/v1/gm/req/sign?id=${this.form.id}&valid_days=${this.form.valid_days}`, "_blank");
            this.$refs['dlg'].hide();
            this.$emit("submit");
        },
        show(id = '', common_name = '') {
            this.errors.clear();
            this.common_name = common_name || '';
            this.form.id = id || '';
            this.$refs['dlg'].show();
        },
    }
}
</script>
