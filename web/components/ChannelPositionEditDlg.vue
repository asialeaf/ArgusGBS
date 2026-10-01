<template>
    <FormDlg title="编辑通道位置" @hide="onHide" @show="onShow" @submit="onSubmit" ref="dlg" :disabled="errors.any() || form.custom">
        <div :class="{'form-group':true,'has-error': errors.has('serial')}">
            <label for="input-position-serial" class="col-sm-4 control-label">设备编号
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" readonly="readonly" id="input-position-serial" name="serial" :value="form.serial" data-vv-as="设备编号" >
            </div>
        </div>
        <div :class="{'form-group':true,'has-error': errors.has('code')}">
            <label for="input-position-code" class="col-sm-4 control-label">通道编号
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" readonly="readonly" id="input-position-code" name="code" :value="form.code" data-vv-as="通道编号">
            </div>
        </div>
        <div :class="{'form-group':true,'has-error': errors.has('name')}">
            <label for="input-position-name" class="col-sm-4 control-label">通道名称
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" readonly="readonly" id="input-position-name" name="name" :value="form.name || '-'" data-vv-as="通道名称">
            </div>
        </div>
        <div :class="{'form-group':true,'has-error': errors.has('longitude')}">
            <label for="input-position-longitude" class="col-sm-4 control-label">经度
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" id="input-position-longitude" name="longitude" v-validate="'decimal'"
                    :placeholder="form.longitude || '-'" v-model.trim="form.customLongitude" data-vv-as="经度"
                    @paste="onPasteLngLat"
                    @keydown.enter="$el.querySelector('#input-position-latitude').focus()">
            </div>
        </div>
        <div :class="{'form-group':true,'has-error': errors.has('latitude')}">
            <label for="input-position-latitude" class="col-sm-4 control-label">纬度
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" id="input-position-latitude" name="latitude" v-validate="'decimal'"
                    :placeholder="form.latitude || '-'" v-model.trim="form.customLatitude" data-vv-as="纬度"
                    @paste="onPasteLngLat"
                    @keydown.enter="onSubmit">
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
            form: this.defForm(),
        }
    },
    components: {
        FormDlg
    },
    methods: {
        defForm() {
            return {
                serial: '',
                code: '',
                name: '',
                custom: false,
                longitude: '',
                latitude: '',
                customLongitude: '',
                customLatitude: '',
            }
        },
        onHide() {
            this.form = this.defForm();
            this.$emit("hide");
        },
        onShow() {
            this.errors.clear();
            this.$emit("show");
            // this.$el.querySelector('#input-position-longitude').focus()
        },
        onPasteLngLat(e) {
            try {
                var vals = e.clipboardData.getData('Text').trim().split(/\s*,\s*/);
                if(vals.length == 2) {
                    this.form.customLongitude = vals[0];
                    this.form.customLatitude = vals[1];
                    e.preventDefault();
                }
            } catch (error) {
                console.log(error);
            }
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
            $.post('/api/v1/device/setchannelposition', {
                serial: this.form.serial,
                code: this.form.code,
                longitude: this.form.customLongitude,
                latitude: this.form.customLatitude,
            }).then(data => {
                this.$refs['dlg'].hide();
                this.$emit("submit");
            })
        },
        show(data) {
            this.errors.clear();
            if(data) {
                this.form.serial = data.serial;
                this.form.code = data.code;
                this.form.name = data.name;
                this.form.longitude = data.longitude || '';
                this.form.latitude = data.latitude || '';
                this.form.customLongitude = data.customLongitude || '';
                this.form.customLatitude = data.customLatitude || '';
            }
            this.$nextTick(() => {
                this.$refs['dlg'].show();
            })
        }
    }
}
</script>
