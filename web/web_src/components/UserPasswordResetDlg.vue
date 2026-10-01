<template>
    <FormDlg :title="title" @hide="onHide" @show="onShow" @submit="onSubmit" ref="dlg" :disabled="errors.any() || (form.newPassword && (matchCount(form.newPassword) < (serverInfo.PwdLevel||3) || form.newPassword.length < (serverInfo.PwdLength||8)))">
        <div :class="{'form-group':true, 'has-feedback':true, 'has-error': errors.has('myPassword')}">
            <label class="col-sm-4 control-label">我的密码
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" name="myPassword" v-model.trim="form.myPassword" data-vv-as="我的密码" v-validate="'required'" @keydown.enter="$el.querySelector('[name=newPassword]').focus()" v-if="bRawMyPwd">
                <input type="password" class="form-control" name="myPassword" v-model.trim="form.myPassword" autocomplete="new-password" data-vv-as="我的密码" v-validate="'required'" @keydown.enter="$el.querySelector('[name=newPassword]').focus()" v-else>
				<span class="glyphicon glyphicon-eye-open form-control-feedback text-gray" @click="bRawMyPwd = !bRawMyPwd;" v-if="bRawMyPwd"></span>
				<span class="glyphicon glyphicon-eye-close form-control-feedback text-gray" @click="bRawMyPwd = !bRawMyPwd;" v-else></span>
            </div>
        </div>
        <div :class="{'form-group':true, 'has-feedback':true, 'has-error': errors.has('newPassword') || (form.newPassword && (matchCount(form.newPassword) < (serverInfo.PwdLevel||3) || form.newPassword.length < (serverInfo.PwdLength||8)))}">
            <label class="col-sm-4 control-label">TA的新密码
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" name="newPassword" v-model.trim="form.newPassword" data-vv-as="TA的新密码" v-validate="'required'" @keydown.enter="onSubmit" v-if="bRawPwd">
                <input type="password" class="form-control" name="newPassword" v-model.trim="form.newPassword" autocomplete="new-password" data-vv-as="TA的新密码" v-validate="'required'" @keydown.enter="onSubmit" v-else>
				<span class="glyphicon glyphicon-eye-open form-control-feedback text-gray" @click="bRawPwd = !bRawPwd;" v-if="bRawPwd"></span>
				<span class="glyphicon glyphicon-eye-close form-control-feedback text-gray" @click="bRawPwd = !bRawPwd;" v-else></span>
                <p class="help-block" v-if="form.newPassword && matchCount(form.newPassword) < (serverInfo.PwdLevel||3)">
                    密码必须包含数字、大小写字母和特殊符号四类中至少{{serverInfo.PwdLevel||3}}类
                </p>
                <p class="help-block" v-else-if="form.newPassword && form.newPassword.length < (serverInfo.PwdLength||8)">
                    密码长度不能小于{{serverInfo.PwdLength||8}}位
                </p>
            </div>
        </div>
    </FormDlg>
</template>

<script>
import FormDlg from "components/FormDlg.vue";
import $ from "jquery";
import moment from "moment";
import crypto from "crypto-js";
import { JSEncrypt } from "jsencrypt";

export default {
    props: {
        serverInfo: {
            type: Object,
            default: () => {}
        },
    },
    data() {
        return {
            form: this.defForm(),
            username: "",
            bRawPwd: true,
            bRawMyPwd: true,
        }
    },
    components: { FormDlg },
    computed: {
        title() {
            if(this.username) {
                return `重置密码(${this.username})`;
            }
            return "重置密码";
        }
    },
    methods: {
        defForm() {
            return {
                uid: '',
                newPassword: '',
                myPassword: '',
            }
        },
        onHide() {
            this.form = this.defForm();
            this.username = "";
        },
        onShow() {
            this.bRawPwd = false;
            this.bRawMyPwd = false;
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
            var password = this.encryptPwd(this.form.newPassword);
            var timestamp = moment().format("YYYYMMDDHHmmss");
            $.post('/api/v1/user/resetpassword', {
                id: this.form.uid,
                password: password,
                timestamp: timestamp,
                verify: this.md5(`${this.form.uid}${password}${timestamp}${this.md5(this.form.myPassword)}`),
            }).then(data => {
                this.$refs['dlg'].hide();
                this.$message({
                    type: 'success',
                    message: "密码重置成功！"
                })
            })
        },
        show(id, username = "") {
            this.bRawPwd = true;
            this.bRawMyPwd = true;
            this.errors.clear();
            this.form.uid = id;
            this.username = username || "";
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
