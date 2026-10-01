<template>
	<div class="login-box">
		<div class="row header-space">
		</div>
		<div class="login-container">
            <div id="login-panel" :class="['panel', 'panel-primary', {'visible': showPanel == 'login'}]">
                <div class="panel-heading text-center">
                    <h3><b>{{serverInfo.LogoText || "LiveGBS"}} <small v-if="serverInfo.DemoUser" style="color:#fff;"><sup><sup>演示</sup></sup></small></b> - 登录</h3>
                </div>
                <div class="panel-body">
                    <br>
                    <form autocomplete="off" id="login-form">
                        <div :class="{'form-group':true, 'has-feedback':true,'has-error': errors.has('username')}">
                            <input type="text" class="form-control" :placeholder="usernamePlaceholder" v-validate="'required'" data-vv-as="用户名" name="username" v-model.trim="username" @keydown.enter="$el.querySelector('#login-form [name=password]').focus()">
                            <span class="glyphicon glyphicon-user form-control-feedback text-gray"></span>
                            <!-- <span class="help-block">{{errors.first('username')}}</span> -->
                        </div>
                        <div :class="{'form-group':true, 'has-feedback':true,'has-error': errors.has('password')}">
                            <input type="text" class="form-control" :placeholder="passwordPlaceholder" v-validate="'required'" data-vv-as="密码" name="password" v-model.trim="password" @keydown.enter="doLogin" v-if="bRawPwd">
                            <input type="password" class="form-control" :placeholder="passwordPlaceholder" autocomplete="new-password" v-validate="'required'" data-vv-as="密码" name="password" v-model.trim="password" @keydown.enter="doLogin" v-else>
                            <span class="glyphicon glyphicon-eye-open form-control-feedback text-gray" @click="bRawPwd = !bRawPwd;" v-if="bRawPwd"></span>
                            <span class="glyphicon glyphicon-eye-close form-control-feedback text-gray" @click="bRawPwd = !bRawPwd;" v-else></span>
                            <!-- <span class="help-block">{{errors.first('password')}}</span> -->
                        </div>
                        <div class="row" v-if="serverInfo.Captcha">
                            <div class="col-xs-7">
                                <div :class="{'form-group':true, 'has-feedback':true,'has-error': errors.has('captcha')}">
                                    <input id="input-captcha" type="text" class="form-control" placeholder="验证码" name="captcha" data-vv-as="验证码" v-model.trim="captcha" v-validate="'required'" @keydown.enter="doLogin">
                                    <span class="glyphicon glyphicon-check form-control-feedback text-gray"></span>
                                </div>
                            </div>
                            <div class="col-xs-5">
                                <img :src="captchaURL" title="验证码" alt="验证码" style="height: 32px;" role="button" @click.prevent="refreshCaptcha">
                            </div>
                        </div>
                        <div class="row">
                            <div class="col-xs-6 col-sm-4" v-if="serverInfo.DemoUser">
                                <button id="btn-annou-login" type="button" class="btn btn-info btn-block btn-flat" @click.prevent="doAnonLogin" :disabled="isLoading">匿名登录</button>
                            </div>
                            <div class="col-xs-6 col-sm-4" v-else>
                                <div class="checkbox" v-if="!serverInfo.HideRememberMe">
                                    <el-checkbox style="margin-top:-5px;padding-left:0;" size="small" v-model.trim="rememberMe">
                                        记住我
                                    </el-checkbox>
                                </div>
                            </div>
                            <div class="col-xs-6 col-sm-offset-4 col-sm-4">
                                <button id="btn-login" type="button" class="btn btn-primary btn-block btn-flat" @click.prevent="doLogin" :disabled="isLoading">登&nbsp;&nbsp;录</button>
                            </div>
                        </div>
                    </form>
                </div>

                <div class="panel-footer">
                    <div class="row">
                        <div class="col-xs-6 text-left">
                            <a class="text-primary" role="button" @click="showPanel = 'forgot'">忘记密码</a>
                        </div>
                        <div class="col-xs-6 text-right" v-if="serverInfo.APIAuth === false && serverInfo.LiveStreamAuth === false">
                            <a href="/" class="text-primary">游客访问</a>
                            <!--<a href="#" class="" data-target="#regist-panel">注 册</a>-->
                        </div>
                    </div>
                </div>
            </div>

            <div id="forgot-panel" :class="['panel', 'panel-danger', {'visible': showPanel == 'forgot'}]">
                <div class="panel-heading">
                    <p class="panel-title text-left"><i class="fa fa-key"></i> 提示</p>
                </div>
                <div class="panel-body text-center">
                    <small>请联系运营人员重置您的密码。</small>
                </div>

                <div class="panel-footer text-center">
                    <a class="text-primary" role="button" @click="showPanel = 'login'">返回登录</a>
                </div>
            </div>

            <div id="demo-tip" class="text-warning text-center" v-if="serverInfo.DemoUser">
                演示系统限制匿名登录有效时间
            </div>
		</div>
	</div>
</template>

<script>
import "font-awesome/css/font-awesome.css";
import "bootstrap/dist/css/bootstrap.css";
import "admin-lte/dist/css/skins/_all-skins.css";
import "assets/styles/AdminLTE-custom.less";
import "assets/styles/custom.less";

import "bootstrap/dist/js/bootstrap.js";
import "admin-lte/dist/js/adminlte.js";

import Vue from "vue";
import { mapState, mapActions } from "vuex";

import VeeValidate, { Validator } from "vee-validate";
import zh_CN from "vee-validate/dist/locale/zh_CN";
Validator.localize("zh_CN", zh_CN);
Vue.use(VeeValidate, {
	locale: "zh_CN",
	delay: 500,
	dictionary: {
		zh_CN: {
			messages: {
				required: field => `${field} 不能为空`,
				confirmed: (field, targetField) => `${field} 和 ${targetField} 不匹配`,
				regex: field => `${field} 不符合要求格式`
			}
		}
	}
});

import ElementUI from "element-ui";
import "assets/styles/element-custom.scss";
import patchElementUI from "elements/patch.js";
Vue.use(ElementUI);
patchElementUI();

import crypto from "crypto-js";
import url from "url";
import $ from "jquery";
import { JSEncrypt } from "jsencrypt";

$.ajaxSetup({ cache: false });
export default {
	data() {
		return {
			captchaURL: `/captcha?_t=${new Date().getTime()}`,
			isLoading: false,
			rememberMe: false,
			bRawPwd: false,
			showPanel: "login",
			username: "",
			password: "",
			captcha: "",
			demoUser: "",
			usernamePlaceholder: "用户名",
			passwordPlaceholder: "密码",
			loginRequestMethod: "POST",
		};
	},
	computed: {
		...mapState(["userInfo", "serverInfo"])
	},
	async mounted() {
		var serverInfo = await this.getServerInfo();
		if(serverInfo) {
			document.title = serverInfo.LogoText || "LiveGBS";
			this.loginRequestMethod = serverInfo.LoginRequestMethod || "POST";
			this.demoUser = serverInfo.DemoUser || "";
			if(serverInfo.Uninitialized === true) {
				this.usernamePlaceholder = "用户名(admin)";
				this.passwordPlaceholder = "密码(admin)";
			}
		}
		var userInfo = await this.getUserInfo();
		if(userInfo) {
            window.location.href = "/";
            return
		}
		this.$el.querySelector("#login-form [name=username]").focus();
	},
	methods: {
		...mapActions([ "getUserInfo", "getServerInfo" ]),
		md5(text) {
			return crypto.MD5(text).toString();
		},
		doAnonLogin() {
            this.isLoading = true;
            var data = {
                username: this.demoUser,
                password: this.encryptPwd(this.md5(this.demoUser)),
            }
            if(this.captcha) {
                data["captcha"] = this.captcha;
            }
            $.ajax({
                method: this.loginRequestMethod,
                url: "/api/v1/login",
                data: data,
            }).then(data => {
                var _url = url.parse(window.location.href, true);
                window.location.href = _url.query.r || "/";
            }).fail(xhr => {
				this.refreshCaptcha();
			}).always(() => {
                this.isLoading = false;
            })
		},
		async doLogin() {
			var ok = await this.$validator.validateAll();
			if (!ok) {
				var e = this.errors.items[0];
				this.$message({
					type: "error",
					message: e.msg
				});
				document.querySelector(`[name=${e.field}]`).focus();
				return;
			}
			this.isLoading = true;
			var data = {
					username: this.username,
					password: this.encryptPwd(this.md5(this.password)),
			}
			if(this.captcha) {
				data["captcha"] = this.captcha;
			}
			if(!this.serverInfo.DemoUser && !this.serverInfo.HideRememberMe) {
				data["remember_me"] = this.rememberMe;
			}
			$.ajax({
				method: this.loginRequestMethod,
				url: "/api/v1/login",
				data: data,
			}).then(data => {
				var _url = url.parse(window.location.href, true);
				if(_url.query.r) {
					var __url = url.parse(_url.query.r, true);
					if(!__url.hostname || __url.hostname == _url.hostname) {
						window.location.href = _url.query.r || "/";
						return
					} else {
						console.log(`ignore redirect cross domain from ${_url.hostname} to ${__url.hostname}`);
					}
				}
				// window.location.href = _url.query.r || "/";
				window.location.href = "/";
			}).fail(xhr => {
				this.refreshCaptcha();
			}).always(() => {
				this.isLoading = false;
			});
		},
		refreshCaptcha() {
			this.captchaURL = `/captcha?_t=${new Date().getTime()}`;
			this.captcha = '';
		},
		encryptPwd(password) {
			if(!this.serverInfo.PwdPub) return password;
			var pubKey = window.atob(this.serverInfo.PwdPub);
			if(pubKey.indexOf("RSA Public Key") < 0) return password;
			var encrypt = new JSEncrypt();
			encrypt.setPublicKey(pubKey);
			return encrypt.encrypt(password);
		},
	}
};
</script>

<style lang="less">
@import url(~assets/styles/variables.less);

.form-group .form-control-feedback {
	&.glyphicon-eye-open, &.glyphicon-eye-close {
		pointer-events: auto;
		cursor: pointer;
	}
}

.light-login {
	background: #dfe0e2 url(~assets/images/pattern.jpg) repeat;
}
.header-space {
	margin: 60px 0;
}

.login-container {
	margin: 0 auto;
	// width: 340px;
}

.panel {
	position: fixed;
	visibility: hidden;
	z-index: auto;

	-moz-transform: scale(0, 1) translate(-150px);
	-webkit-transform: scale(0, 1) translate(-150px);
	-o-transform: scale(0, 1) translate(-150px);
	-ms-transform: scale(0, 1) translate(-150px);
	transform: scale(0, 1) translate(-150px);
}

.panel-primary {
	border-color: @base;
}

.panel-primary > .panel-heading {
    color: #fff;
    background-color: @base;
    border-color: @base;
}

.panel.visible {
	visibility: visible;
	position: relative;

	-moz-transform: scale(1, 1) translate(0);
	-webkit-transform: scale(1, 1) translate(0);
	-o-transform: scale(1, 1) translate(0);
	-ms-transform: scale(1, 1) translate(0);
	transform: scale(1, 1) translate(0);

	transition: transform 0.3s ease;
	-moz-transition: -moz-transform 0.3s ease;
	-webkit-transition: -webkit-transform 0.3s ease;
	-o-transition: -o-transform 0.2s ease;
}

@media (max-width: 768px) {
	.login-box, .register-box {
		width: 90%;
	}
}
</style>
