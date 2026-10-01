<template>
    <FormDlg title="备注证书请求" @hide="onHide" @show="onShow" @submit="onSubmit" ref="dlg" :disabled="errors.any()">
        <div :class="{'form-group':true,'has-error': errors.has('serial')}">
            <label for="input-serial" class="col-sm-4 control-label">序列号
                <!-- <span class="text-red">*</span> -->
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" readonly="readonly" id="input-serial" name="serial" :value="form.serial" data-vv-as="序列号" >
            </div>
        </div>
        <div :class="{'form-group':true,'has-error': errors.has('common_name')}">
            <label for="input-common-name" class="col-sm-4 control-label">签发给
                <!-- <span class="text-red">*</span> -->
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" readonly="readonly" id="input-common-name" name="common_name" :value="form.common_name" data-vv-as="签发给">
            </div>
        </div>
        <div :class="{'form-group':true,'has-error': errors.has('description')}">
            <label for="input-description" class="col-sm-4 control-label">备注
            </label>
            <div class="col-sm-7">
                <el-input id="input-description" type="textarea" v-model.trim="form.description" :autosize="{minRows:3, maxRows:10}" :rows="3" placeholder="可选"></el-input>
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
            form: this.defForm(),
        }
    },
    components: {
        FormDlg
    },
    methods: {
        defForm() {
            return {
                id: '',
                serial: '',
                common_name: '',
                description: '',
            }
        },
        onHide() {
            this.form = this.defForm();
            this.$emit("hide");
        },
        onShow() {
            this.errors.clear();
            this.$emit("show");
            this.$el.querySelector('#input-description').focus();
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
            $.post('/api/v1/gm/req/setdescription', {
                id: this.form.id,
                description: this.form.description,
            }).then(data => {
                this.$refs['dlg'].hide();
                this.$emit("submit");
            })
        },
        show(data) {
            this.errors.clear();
            if(data) {
                this.form.id = data.id;
                this.form.serial = data.serial;
                this.form.common_name = data.common_name;
                this.form.description = data.description;
            }
            this.$nextTick(() => {
                this.$refs['dlg'].show();
            })
        }
    }
}
</script>
