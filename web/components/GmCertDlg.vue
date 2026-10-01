<template>
    <FormDlg title="创建自签名国密证书" ref="dlg" @hide="onHide" @show="onShow" @submit="onSubmit" :disabled="errors.any()">
        <div :class="{'form-group': true, 'has-feedback': true,'has-error': errors.has('valid_days')}">
            <label class="col-sm-4 control-label">有效期(天)
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" name="valid_days"
                    v-model.trim="form.valid_days" data-vv-as="有效期" v-validate="'required|numeric|min_value:1'" @keydown.enter.prevent="onSubmit">
            </div>
        </div>
        <div :class="{'form-group': true, 'has-feedback': true,'has-error': errors.has('ca')}" v-show="false">
            <label class="col-sm-4 control-label">其它配置</label>
            <div class="col-sm-7 checkbox">
                <el-checkbox style="margin-left:-19px;margin-top:-5px;" size="small" v-model.trim="form.ca" name="ca">
                    CA&nbsp;&nbsp;
                </el-checkbox>
            </div>
        </div>
    </FormDlg>
</template>

<script>
import FormDlg from 'components/FormDlg.vue'
import $ from 'jquery'

export default {
    props: {
        action: {
            default: '/api/v1/gm/cert/create',
        }
    },
    data() {
        return {
            form: this.defForm(),
        }
    },
    components: {
        FormDlg
    },
    methods: {
        defForm() {
            return {
                valid_days: 3650,
                req: "",
                ca: true,
            }
        },
        onHide() {
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
            $.post(this.action, this.form).then(() => {
                this.$refs['dlg'].hide();
                this.$emit("submit");
            })
        },
        show(req = '') {
            this.errors.clear();
            this.form.req = req || '';
            this.$refs['dlg'].show();
        },
    }
}
</script>
