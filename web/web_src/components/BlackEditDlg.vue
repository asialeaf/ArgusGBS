<template>
<FormDlg title="编辑黑名单" @hide="onHide" @show="onShow" @submit="onSubmit" ref="dlg" :disabled="errors.any()">
    <div :class="{'form-group':true,'has-error': errors.has('serial')}">
        <label class="col-sm-4 control-label">设备国标编号
            <span class="text-red">*</span>
        </label>
        <div class="col-sm-7">
            <input type="text" class="form-control" :readonly="form.oper == 'edit'" name="serial" v-model.trim="form.serial" data-vv-as="设备编号" v-validate="'required'">
        </div>
    </div>
    <div :class="{'form-group':true,'has-error': errors.has('user_agent')}">
        <label class="col-sm-4 control-label">UA
        </label>
        <div class="col-sm-7">
            <input type="text" class="form-control" name="user_agent" v-model.trim="form.user_agent" placeholder="为空时不限接入UA" data-vv-as="UA">
        </div>
    </div>
    <div :class="{'form-group':true,'has-error': errors.has('remote_ip')}">
        <label class="col-sm-4 control-label">IP
        </label>
        <div class="col-sm-7">
            <input type="text" class="form-control" name="remote_ip" v-model.trim="form.remote_ip" placeholder="为空时不限接入IP" data-vv-as="IP">
        </div>
    </div>
    <div :class="{'form-group':true,'has-error': errors.has('remote_port')}">
        <label class="col-sm-4 control-label">端口
        </label>
        <div class="col-sm-7">
            <input type="text" class="form-control" name="remote_port" v-model.trim="form.remote_port" placeholder="为空时不限接入端口" v-validate="'numeric'" data-vv-as="端口">
        </div>
    </div>
    <div :class="{'form-group':true,'has-error': errors.has('description')}">
        <label class="col-sm-4 control-label">备注
        </label>
        <div class="col-sm-7">
            <input type="text" class="form-control" name="description" v-model.trim="form.description" data-vv-as="描述信息">
        </div>
    </div>
</FormDlg>
</template>

<script>
import FormDlg from 'components/FormDlg.vue'
import $ from 'jquery'

export default {
    data() {
        return {
            form: this.defForm()
        }
    },
    components: {
        FormDlg
    },
    methods: {
        defForm() {
            return {
                oper : 'edit',
                serial: '',
                description: '',
                user_agent: '',
                remote_ip: '',
                remote_port: '',
            }
        },
        onHide() {
            this.form = this.defForm();
            this.$emit("hide");
        },
        onShow() {
            this.errors.clear();
            this.$emit("show");
            if (!this.form.serial) {
                this.form.oper = 'add';
                this.$el.querySelector('[name=serial]').focus();
            }
        },
        async onSubmit() {
            var ok = await this.$validator.validateAll();
            if (!ok) {
                var e = this.errors.items[0];
                this.$message({
                    type: 'error',
                    message: e.msg
                })
                $(`[name=${e.field}]`).focus();
                return;
            }
            $.post('/api/v1/black/save', this.form).then(data => {
                this.$refs['dlg'].hide();
                this.$emit("submit");
            })
        },
        show(data) {
            this.errors.clear();
            if (data) {
                Object.assign(this.form, data);
            }
            this.$nextTick(() => {
                this.$refs['dlg'].show();
            })
        }
    }
}
</script>
