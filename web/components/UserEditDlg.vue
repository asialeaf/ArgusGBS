<template>
    <FormDlg title="编辑用户" @hide="onHide" @show="onShow" @submit="onSubmit" ref="dlg" :disabled="errors.any()">
        <input type="hidden" name="ID" v-model.trim="form.ID">
        <div :class="{'form-group':true, 'has-error': errors.has('Username')}">
            <label for="input-username" class="col-sm-4 control-label">用户名
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" id="input-username" name="Username" v-model.trim="form.Username" :readonly="reserve" data-vv-as="用户名" v-validate="'required'" @keydown.enter.prevent="$el.querySelector('#input-role').focus()">
            </div>
        </div>
        <div :class="{'form-group':true, 'has-error': errors.has('Role')}">
            <label for="input-role" class="col-sm-4 control-label">角色
                <!-- <span class="text-red">*</span> -->
            </label>
            <div class="col-sm-7">
                <el-select id="input-role" style="width:100%;" size="medium" v-model.trim="roles" :disabled="reserve" multiple filterable allow-create default-first-option placeholder="请选择">
                    <el-option v-for="(item, idx) in innerRoles" :key="idx" :label="item" :value="item">
                    </el-option>
                </el-select>
                <!-- <input type="text" class="form-control" id="input-role" name="Role" v-model.trim="form.Role" data-vv-as="角色" @keydown.enter="onSubmit"> -->
            </div>
        </div>
        <div :class="{'form-group':true, 'has-error': errors.has('ControlPriority')}" v-show="roles && roles.some(r => r == '操作员' || r == '管理员' || r == '超级管理员')">
            <label for="input-control-priority" class="col-sm-4 control-label">云台控制优先级</label>
            <div class="col-sm-7">
                <el-slider id="input-control-priority" name="ControlPriority" data-vv-as="云台控制优先级" v-model.number="form.ControlPriority" :min="0" :max="100"></el-slider>
            </div>
        </div>
        <div :class="{'form-group':true, 'has-error': errors.has('PhoneNumber')}">
            <label for="input-phone-number" class="col-sm-4 control-label">手机号
                <!-- <span class="text-red">*</span> -->
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" id="input-phone-number" name="PhoneNumber" v-model.trim="form.PhoneNumber" data-vv-as="手机号" v-validate="'regex:^1\\d{10}$'" placeholder="可选" @keydown.enter.prevent="$el.querySelector('#input-email').focus()">
            </div>
        </div>
        <div :class="{'form-group':true, 'has-error': errors.has('Email')}">
            <label for="input-email" class="col-sm-4 control-label">邮箱
                <!-- <span class="text-red">*</span> -->
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" id="input-email" name="Email" v-model.trim="form.Email" data-vv-as="邮箱" v-validate="'email'" placeholder="可选" @keydown.enter.prevent="$el.querySelector('#input-description').focus()">
            </div>
        </div>
        <div :class="{'form-group':true,'has-error': errors.has('Description')}">
            <label for="input-description" class="col-sm-4 control-label">备注
            </label>
            <div class="col-sm-7">
                <el-input id="input-description" type="textarea" v-model.trim="form.Description" :autosize="{minRows:1, maxRows:10}" :rows="1" placeholder="可选"></el-input>
            </div>
        </div>
        <div :class="{'form-group':true, 'has-error': errors.has('Enable')}" v-if="!reserve">
            <label class="col-sm-4 control-label">其它选项
                <!-- <span class="text-red">*</span> -->
            </label>
            <div class="col-sm-7 checkbox">
                <el-checkbox style="margin-left:-19px;margin-top:-5px;" size="small" v-model.trim="form.Enable" name="Enable">
                    启用
                </el-checkbox>
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
            innerRoles: ['管理员', '操作员', '观众'],
            roles: [],
            reserve: false,
        }
    },
    components: {
        FormDlg
    },
    methods: {
        defForm() {
            return {
                ID: "",
                Username: "",
                Role: "",
                ControlPriority: 0,
                PhoneNumber: "",
                Email: "",
                Description: "",
                Enable: true,
            }
        },
        onHide() {
            this.form = this.defForm();
            this.roles = [];
            this.reserve = false;
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
            this.form.Role = this.roles.join(",");
            $.post('/api/v1/user/save', this.form).then(data => {
                if(!this.form.ID) {
                    this.$message({
                        type: 'success',
                        duration: 15000,
                        showClose: true,
                        message: `创建用户成功, 初始密码 ${data.DefaultUserPassword}`,
                    })
                } else {
                    this.$message({
                        type: 'success',
                        message: `编辑用户成功`,
                    })
                }
                this.$refs['dlg'].hide();
                this.$emit("submit");
            })
        },
        show(data, reserve = false) {
            this.errors.clear();
            if(data) {
                Object.assign(this.form, data);
            }
            if(this.form.Role) {
                this.roles = this.form.Role.split(",");
            }
            this.reserve = reserve;
            this.$nextTick(() => {
                this.$refs['dlg'].show();
            })
        }
    }
}
</script>
