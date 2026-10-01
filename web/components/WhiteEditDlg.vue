<template>
<FormDlg title="编辑白名单" @hide="onHide" @show="onShow" @submit="onSubmit" ref="dlg" :disabled="errors.any() || bCertUploading" v-loading="bCertUploading">
    <div :class="{'form-group':true, 'has-error': errors.has('serial')}">
        <label class="col-sm-4 control-label">设备国标编号
            <span class="text-red">*</span>
        </label>
        <div class="col-sm-7">
            <input type="text" class="form-control" :readonly="form.oper == 'edit'" name="serial" v-model.trim="form.serial" data-vv-as="设备编号" v-validate="'required'">
        </div>
    </div>
    <div :class="{'form-group':true, 'has-error': errors.has('direction')}" v-show="showDirection">
        <label class="col-sm-4 control-label">认证方式
        </label>
        <div class="col-sm-7">
            <select class="form-control" id="input-direction" name="direction" v-model.trim.number="form.direction" data-vv-as="认证方式" v-validate="">
                <option value="0">28181 摘要</option>
                <option value="1">35114 单向</option>
                <option value="2">35114 双向</option>
            </select>
        </div>
    </div>
    <div :class="{'form-group':true, 'has-feedback': false, 'has-error': errors.has('password')}" v-show="!form.direction">
        <label class="col-sm-4 control-label">接入密码
        </label>
        <div class="col-sm-7">
            <input type="text" class="form-control" name="password" v-model.trim="form.password" placeholder="为空时免密接入" data-vv-as="接入密码">
            <!-- <input type="text" class="form-control" name="password" v-model.trim="form.password" placeholder="为空时免密接入" data-vv-as="接入密码" v-if="bRawPwd">
            <input type="password" class="form-control" name="password" v-model.trim="form.password" placeholder="为空时免密接入" autocomplete="new-password" data-vv-as="接入密码" v-else>
            <span class="glyphicon glyphicon-eye-open form-control-feedback text-gray" @click="bRawPwd = !bRawPwd;" v-if="bRawPwd"></span>
            <span class="glyphicon glyphicon-eye-close form-control-feedback text-gray" @click="bRawPwd = !bRawPwd;" v-else></span> -->
        </div>
    </div>
    <div :class="{'form-group':true, 'has-error': errors.has('cert')}" v-show="form.cert || form.direction">
        <label class="col-sm-4 control-label">设备证书
        </label>
        <div class="col-sm-7 checkbox">
            <a :href="`/api/v1/gm/cert/download?serial=${form.serial}`" target="_blank" title="下载设备国密证书" v-if="form.cert">
                <i class="fa fa-download"></i> 下载证书&nbsp;&nbsp;
            </a>
            <a role="button" @click="removeCert" title="删除设备国密证书" v-if="form.cert">
                <i class="fa fa-remove"></i> 删除证书
            </a>
            <el-upload :action="`/api/v1/gm/cert/import?serial=${form.serial}`" accept=".pem,.cer,.crt,.csr,.der" v-if="!form.cert"
                style="display:inline-block;margin-right:7px;"
                :show-file-list="false"
                :on-success="onUploadCertSuccess" :on-error="onUploadCertError" :on-progress="onUploadCertProgress">
                <a role="button" title="上传设备国密证书">
                    <i class="fa fa-upload"></i> 上传设备证书
                </a>
            </el-upload>
        </div>
    </div>
    <div :class="{'form-group':true, 'has-error': errors.has('user_agent')}">
        <label class="col-sm-4 control-label">UA
        </label>
        <div class="col-sm-7">
            <input type="text" class="form-control" name="user_agent" v-model.trim="form.user_agent" placeholder="为空时不限接入UA" data-vv-as="UA">
        </div>
    </div>
    <div :class="{'form-group':true, 'has-error': errors.has('remote_ip')}">
        <label class="col-sm-4 control-label">IP
        </label>
        <div class="col-sm-7">
            <input type="text" class="form-control" name="remote_ip" v-model.trim="form.remote_ip" placeholder="为空时不限接入IP" data-vv-as="IP">
        </div>
    </div>
    <div :class="{'form-group':true, 'has-error': errors.has('remote_port')}">
        <label class="col-sm-4 control-label">端口
        </label>
        <div class="col-sm-7">
            <input type="text" class="form-control" name="remote_port" v-model.trim="form.remote_port" placeholder="为空时不限接入端口" v-validate="'numeric'" data-vv-as="端口">
        </div>
    </div>
    <div :class="{'form-group':true, 'has-error': errors.has('description')}">
        <label class="col-sm-4 control-label">备注
        </label>
        <div class="col-sm-7">
            <!-- <input type="text" class="form-control" name="description" v-model.trim="form.description" data-vv-as="描述信息"> -->
            <el-input id="input-description" type="textarea" v-model.trim="form.description" :autosize="{minRows:1, maxRows:5}" :rows="1" data-vv-as="描述信息"></el-input>
        </div>
    </div>
    <div :class="{'form-group':true, 'has-error': errors.has('save_device')}">
        <label for="input-save-device" class="col-sm-4 control-label">其他选项
        </label>
        <div class="col-sm-7 checkbox">
            <el-checkbox style="margin-left:-19px;margin-top:-5px;" size="small" v-model.trim="form.save_device" name="save_device">
                加到设备列表
            </el-checkbox>
        </div>
    </div>
</FormDlg>
</template>

<script>
import FormDlg from 'components/FormDlg.vue';
import $ from 'jquery';

export default {
    props: {
		serverInfo: {
			type: Object,
			default: () => {}
		},
		userInfo: {
			type: Object,
			default: () => null
		}
    },
    data() {
        return {
            form: this.defForm(),
            debug: false,
            bRawPwd: true,
            bCertUploading: false,
            showDirection: false,
        }
    },
    components: {
        FormDlg
    },
    beforeDestroy() {
        $(this.$el).off("keydown", this.keyDown);
    },
    methods: {
        defForm() {
            return {
                oper : 'edit',
                serial: '',
                description: '',
                password: '',
                user_agent: '',
                remote_ip: '',
                remote_port: '',
                save_device: false,
                direction: 0,
                cert: '',
            }
        },
        onHide() {
            this.debug = false;
            this.bCertUploading = false;
            this.showDirection = false;
            this.form = this.defForm();
            this.$emit("hide");
            $(this.$el).off("keydown", this.keyDown);
        },
        onShow() {
            this.bRawPwd = false;
            this.errors.clear();
            this.$emit("show");
            if (!this.form.serial) {
                this.form.oper = 'add';
                this.$el.querySelector('[name=serial]').focus();
            }
            $(this.$el).on("keydown", this.keyDown);
        },
        keyDown(e) {
            if(e.altKey && e.shiftKey) {
                switch(e.key) {
                    case 'D':
                        e.preventDefault();
                        this.toggleDebug();
                        break;
                }
            }
        },
        toggleDebug() {
            this.debug = !this.debug;
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
            $.post('/api/v1/white/save', this.form).then(data => {
                this.$refs['dlg'].hide();
                this.$emit("submit");
            })
        },
        show(data) {
            this.bRawPwd = true;
            this.errors.clear();
            if (data) {
                Object.assign(this.form, data);
                if (this.form.remote_port === 0) {
                    this.form.remote_port = '';
                }
            }
            this.showDirection = (this.serverInfo.GM || this.form.direction);
            this.$nextTick(() => {
                this.$refs['dlg'].show();
            })
        },
        onUploadCertSuccess(res, file, fileList) {
            this.bCertUploading = false;
            this.form.cert = res;
        },
        onUploadCertProgress(evt, file, fileList) {
            this.bCertUploading = true;
        },
        onUploadCertError(err, file, fileList) {
            this.bCertUploading = false;
            if (err) {
                this.$message({
                    type: "error",
                    message: err + ""
                })
            }
        },
        removeCert() {
            $(this.$el).hide();
            this.$confirm("确认删除设备证书?", "提示", {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
                lockScroll: false,
            }).then(() => {
                $(this.$el).show().focus();
                $.post(`/api/v1/gm/cert/remove?serial=${this.form.serial}`).then(() => {
                    this.form.cert = "";
                })
            }).catch(() => {
                $(this.$el).show().focus();
            });
        },
    }, //-- methods
}
</script>

<style lang="less" scoped>
.form-group .form-control-feedback {
	&.glyphicon-eye-open, &.glyphicon-eye-close {
		pointer-events: auto;
		cursor: pointer;
	}
}
</style>
