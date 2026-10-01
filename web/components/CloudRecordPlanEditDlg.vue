<template>
<FormDlg title="编辑录像计划" @hide="onHide" @show="onShow" @submit="onSubmit" ref="dlg" :disabled="errors.any()" size="modal-lgg" tabindex="">
    <input type="hidden" name="ID" v-model.trim="form.ID">
    <div class="col-md-12">
        <div :class="{'form-group':true, 'has-error': errors.has('Name')}">
            <label for="name" class="col-sm-2 control-label">名称
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-9">
                <input type="text" class="form-control" id="name" name="Name" v-model.trim="form.Name" data-vv-as="名称" v-validate="'required'" @keydown.enter.prevent>
            </div>
        </div>

        <div :class="{'form-group':true, 'has-error': errors.has('Enable')}">
            <label class="col-sm-2 control-label">状态
            </label>
            <div class="col-sm-9 checkbox">
                <el-checkbox style="margin-left:-19px;margin-top:-5px;" size="small" v-model.trim="form.Enable" name="Enable">
                    启用&nbsp;&nbsp;
                </el-checkbox>
                <el-checkbox style="margin-left:-19px;margin-top:-5px;" size="small" v-model.trim="form.Supply" name="Supply">
                    补录&nbsp;&nbsp;
                </el-checkbox>
            </div>
        </div>
        <div class="form-group" v-if="form.Supply">
            <label class="col-sm-2 control-label"></label>
            <div class="col-sm-9" style="border:1px dashed #ccc;">
                <div :class="{'form-group':true, 'has-error': errors.has('SupplyDays')}">
                    <label class="col-sm-2 control-label" style="margin-top:23px;">补录天数(天)
                    </label>
                    <div class="col-sm-9">
                        <span style="color:lightgrey;font-style:italic;line-height:24px;">补录的设备端需要有设备录像 可在 国标设备-》查看通道-》设备录像 查看</span>
                        <input type="text" class="form-control" id="input-supply-days" name="SupplyDays" v-model.trim="form.SupplyDays" v-validate="'numeric'" placeholder="3" @keydown.enter="$el.querySelector('#input-supply-days').focus()">
                    </div>
                </div>
                <div :class="{'form-group':true, 'has-error': errors.has('SupplySpeed')}">
                    <label class="col-sm-2 control-label" style="margin-top:23px;">补录流传输倍速
                    </label>
                    <div class="col-sm-9">
                        <span style="color:lightgrey;font-style:italic;line-height:24px;">倍速不宜配置过大，会消耗带宽 或 设备无法支持</span>
                        <input type="text" class="form-control" id="input-supply-speed" name="SupplySpeed" v-model.trim="form.SupplySpeed" v-validate="'numeric'" placeholder="1" @keydown.enter="$el.querySelector('#input-supply-speed').focus()">
                    </div>
                </div>
                <div :class="{'form-group':true, 'has-error': errors.has('SupplySpeed')}">
                    <label class="col-sm-2 control-label" style="margin-top:23px;">补录执行时间段
                    </label>
                    <div class="col-sm-9">
                        <span style="color:lightgrey;font-style:italic;line-height:24px;">不配置，默认是全天有效</span><br />
                        <el-time-picker value-format="HH:mm:ss" v-model="form.SupplyValidStart" placeholder="开始时间"></el-time-picker>
                        &nbsp;&nbsp;至&nbsp;&nbsp;
                        <el-time-picker value-format="HH:mm:ss" v-model="form.SupplyValidEnd" placeholder="结束时间"></el-time-picker>
                    </div>
                </div>
            </div>
        </div>
        <div :class="{'form-group':true}">
            <label for="name" class="col-sm-2 control-label">录像计划详情
                <!-- <span class="text-red">*</span> -->
            </label>
            <div class="col-sm-9">
                <CloudRecordPlan ref="recordPlan"></CloudRecordPlan>
            </div>
        </div>
    </div>
    <div class="clearfix"></div>
</FormDlg>
</template>

<script>
import FormDlg from 'components/FormDlg.vue'
import CloudRecordPlan from 'components/CloudRecordPlan.vue'
import $ from 'jquery'

export default {
    data() {
        return {
            form: this.defForm()
        }
    },
    components: {
        FormDlg,
        CloudRecordPlan
    },
    methods: {
        defForm() {
            return {
                ID: "",
                Name: "",
                Enable: false,
                Plan: "",
                Supply: false,
                SupplyDays: 3,
                SupplySpeed: 1,
                SupplyValidStart: "",
                SupplyValidEnd: "",
            }
        },
        onHide() {
            this.form = this.defForm();
        },
        onShow() {
            this.errors.clear();
            this.$refs["recordPlan"].init(this.form.Plan);
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
            var params = {
                ID: this.form.ID,
                Name: this.form.Name,
                Plan: this.$refs["recordPlan"].getplan(),
                Enable: this.form.Enable,
            }
            if (this.form.Supply) {
                params["Supply"] = this.form.Supply;
                params["SupplyDays"] = this.form.SupplyDays;
                params["SupplySpeed"] = this.form.SupplySpeed;
                params["SupplyValidStart"] = this.form.SupplyValidStart;
                params["SupplyValidEnd"] = this.form.SupplyValidEnd;
            }

            $.post('/api/v1/cloudrecord/plan/save', params).then(data => {
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

<style lang="less" scoped>
.model-lg {
    width: 80% !important;
}
#name {
    max-width: 800px;
}
</style>
