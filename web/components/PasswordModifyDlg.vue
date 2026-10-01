<template>
    <FormDlg :title="title" @hide="onHide" @show="onShow" @submit="onSubmit" ref="dlg" :disabled="errors.any() || (form.newPassword && (matchCount(form.newPassword) < (serverInfo.PwdLevel||3) || form.newPassword.length < (serverInfo.PwdLength||8)))" :closeable="!userInfo || !userInfo.PwdModReq">
        <div :class="{'form-group':true, 'has-feedback':true, 'has-error': errors.has('oldPassword')}" v-if="!userInfo || !userInfo.PwdModReq">
            <label for="old-password" class="col-sm-4 control-label">原密码
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" id="old-password" name="oldPassword" v-model.trim="form.oldPassword" data-vv-as="原密码" v-validate="'required'" @keydown.enter="$el.querySelector('#new-password').focus()" v-if="bRawOldPwd">
                <input type="password" class="form-control" id="old-password" name="oldPassword" v-model.trim="form.oldPassword" autocomplete="new-password" data-vv-as="原密码" v-validate="'required'" @keydown.enter="$el.querySelector('#new-password').focus()" v-else>
				<span class="glyphicon glyphicon-eye-open form-control-feedback text-gray" @click="bRawOldPwd = !bRawOldPwd;" v-if="bRawOldPwd"></span>
				<span class="glyphicon glyphicon-eye-close form-control-feedback text-gray" @click="bRawOldPwd = !bRawOldPwd;" v-else></span>
            </div>
        </div>
        <div :class="{'form-group':true, 'has-feedback':true, 'has-error': errors.has('newPassword') || (form.newPassword && (matchCount(form.newPassword) < (serverInfo.PwdLevel||3) || form.newPassword.length < (serverInfo.PwdLength||8)))}">
            <label for="new-password" class="col-sm-4 control-label">新密码
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" id="new-password" name="newPassword" v-model.trim="form.newPassword" data-vv-as="新密码" v-validate="'required'" @keydown.enter="$el.querySelector('#new-password2').focus()" v-if="bRawNewPwd">
                <input type="password" class="form-control" id="new-password" name="newPassword" v-model.trim="form.newPassword" autocomplete="new-password" data-vv-as="新密码" v-validate="'required'" @keydown.enter="$el.querySelector('#new-password2').focus()" v-else>
				<span class="glyphicon glyphicon-eye-open form-control-feedback text-gray" @click="bRawNewPwd = !bRawNewPwd;" v-if="bRawNewPwd"></span>
				<span class="glyphicon glyphicon-eye-close form-control-feedback text-gray" @click="bRawNewPwd = !bRawNewPwd;" v-else></span>
                <p class="help-block" v-if="form.newPassword && matchCount(form.newPassword) < (serverInfo.PwdLevel||3)">
                    密码必须包含数字、大小写字母和特殊符号四类中至少{{serverInfo.PwdLevel||3}}类
                </p>
                <p class="help-block" v-else-if="form.newPassword && form.newPassword.length < (serverInfo.PwdLength||8)">
                    密码长度不能小于{{serverInfo.PwdLength||8}}位
                </p>
            </div>
        </div>
        <div :class="{'form-group':true, 'has-feedback':true, 'has-error': errors.has('newPassword2')}">
            <label for="new-password2" class="col-sm-4 control-label">确认密码
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" id="new-password2" name="newPassword2" v-model.trim="form.newPassword2" data-vv-as="确认密码" v-validate="'required|confirmed:newPassword'" @keydown.enter="onSubmit" v-if="bRawNewPwd2">
                <input type="password" class="form-control" id="new-password2" name="newPassword2" v-model.trim="form.newPassword2" autocomplete="new-password" data-vv-as="确认密码" v-validate="'required|confirmed:newPassword'" @keydown.enter="onSubmit" v-else>
				<span class="glyphicon glyphicon-eye-open form-control-feedback text-gray" @click="bRawNewPwd2 = !bRawNewPwd2;" v-if="bRawNewPwd2"></span>
				<span class="glyphicon glyphicon-eye-close form-control-feedback text-gray" @click="bRawNewPwd2 = !bRawNewPwd2;" v-else></span>
            </div>
        </div>
        <template slot="footer" v-if="userInfo && userInfo.PwdModReq">
            <a role="button" class="btn btn-default" href="/logout">退出登录</a>
        </template>
    </FormDlg>
</template>

<script>
import FormDlg from "components/FormDlg.vue";
import $ from "jquery";
import crypto from "crypto-js";
import { JSEncrypt } from "jsencrypt";

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
            bRawOldPwd: true,
            bRawNewPwd: true,
            bRawNewPwd2: true,
        }
    },
    components: { FormDlg },
    computed: {
        title() {
            if(this.userInfo && this.userInfo.PwdModReq) {
                if(this.userInfo.PwdExpDays > 0) {
                    return "密码已过期, 请修改密码";
                }
                return "设置安全登录密码";
            }
            return "修改密码";
        },
    },
    methods: {
        defForm() {
            return {
                oldPassword: '',
                newPassword: '',
                newPassword2: ''
            }
        },
        onHide() {
            this.form = this.defForm();
        },
        onShow() {
            this.bRawOldPwd = false;
            this.bRawNewPwd = false;
            this.bRawNewPwd2 = false;
            this.errors.clear();
            // this.$el.querySelector('#old-password').focus();
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
            if(this.matchCount(this.form.newPassword) < (this.serverInfo.PwdLevel || 3)) {
                this.$message({
                    type: 'error',
                    message: `密码必须包含数字、大小写字母和特殊符号四类中至少${this.serverInfo.PwdLevel || 3}类`
                });
                return;
            }
            if(this.form.newPassword.length < (this.serverInfo.PwdLength || 8)) {
                this.$message({
                    type: 'error',
                    message: `密码长度不能小于${this.serverInfo.PwdLength || 8}位`
                });
                return;
            }
            if(this.form.newPassword === this.form.oldPassword) {
                this.$message({
                    type: 'error',
                    message: `新密码不能与原密码相同`
                });
                return;
            }
            $.post('/api/v1/modifypassword', {
                oldpassword: this.encryptPwd(this.md5(this.form.oldPassword)),
                newpassword: this.encryptPwd(this.md5(this.form.newPassword))
            }).then(data => {
                this.$refs['dlg'].hide();
                this.$alert("密码修改成功,即将重新登录!", "提示", {
                    lockScroll: false,
                }).then(() => {
                    window.location.href = "/logout";
                }).catch(() => {
                    window.location.href = "/logout";
                })
            })
        },
        show(data) {
            this.bRawOldPwd = true;
            this.bRawNewPwd = true;
            this.bRawNewPwd2 = true;
            this.errors.clear();
            if(data) {
                Object.assign(this.form, data);
            }
            this.$refs['dlg'].show();
        },
        md5(text) {
            return crypto.MD5(text).toString();
        },
        encryptPwd(password) {
            if(!this.serverInfo.PwdPub) return password;
            var pubKey = window.atob(this.serverInfo.PwdPub);
            if(pubKey.indexOf("RSA Public Key") < 0) return password;
            var encrypt = new JSEncrypt();
            encrypt.setPublicKey(pubKey);
            return encrypt.encrypt(password);
        },
        matchCount(password) {
            var ret = 0
            if(/\d/.test(password)) ret++; // 数字
            if(/[a-z]/.test(password)) ret++ // 小写
            if(/[A-Z]/.test(password)) ret++; // 大写
            if(/[\W_]/.test(password)) ret++; // 特殊字符
            return ret;
        },
    }
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
