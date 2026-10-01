<template>
    <FormDlg :title="title" @hide="onHide" @show="onShow" @submit="onSubmit" ref="dlg" :disabled="errors.any()">
        <div :class="{'form-group':true,'has-error': errors.has('serial')}">
            <label for="share-node-serial" class="col-sm-4 control-label">设备编号
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" readonly="readonly" id="share-node-serial" name="serial" :value="form.serial" data-vv-as="设备编号" >
            </div>
        </div>
        <div :class="{'form-group':true,'has-error': errors.has('code')}">
            <label for="share-node-code" class="col-sm-4 control-label">节点编号
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" readonly="readonly" id="share-node-code" name="code" :value="form.code" data-vv-as="通道编号">
            </div>
        </div>
        <div :class="{'form-group':true,'has-error': errors.has('name')}">
            <label for="share-node-name" class="col-sm-4 control-label">节点名称
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" readonly="readonly" id="share-node-name" name="name" :value="form.name" data-vv-as="节点名称">
            </div>
        </div>
        <div :class="{'form-group':true, 'has-error': errors.has('target')}">
            <label class="col-sm-4 control-label">共享给
            </label>
            <div class="col-sm-7 checkbox">
                <el-radio style="margin-left:-19px;margin-top:-5px;" size="small" v-model="form.target" label="cascade" v-show="vtype != '标准版'">上级平台&nbsp;&nbsp;&nbsp;&nbsp;</el-radio>
                <el-radio style="margin-left:-19px;margin-top:-5px;" size="small" v-model="form.target" label="user">用 户</el-radio>
            </div>
        </div>
        <div :class="{'form-group':true,'has-error': errors.has('cascade_id')}" v-if="form.target == 'cascade'">
            <label for="share-node-cascade-id" class="col-sm-4 control-label">上级平台
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <el-select filterable clearable remote :remote-method="loadCascade" :loading="bLoading" v-model="form.cascade_id"
                    @change="changeCascade" @focus="!form.cascade_id && loadCascade();" ref="selectCascade"
                    id="share-node-cascade-id" name="cascade_id"
                    placeholder="可搜索" size="small" style="width:100%;">
                    <el-option v-for="item in cascades" :key="item.ID" :label="renderCascadeName(item)" :value="item.ID">
                    </el-option>
                </el-select>
            </div>
        </div>
        <div :class="{'form-group':true,'has-error': errors.has('username')}" v-if="form.target == 'user'">
            <label for="share-node-username" class="col-sm-4 control-label">用 户
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <el-select filterable clearable remote :remote-method="loadUser" :loading="bLoading" v-model="form.username"
                    @change="changeUser" @focus="!form.username && loadUser();" ref="selectUser"
                    id="share-node-username" name="username"
                    placeholder="可搜索" size="small" style="width:100%;">
                    <el-option v-for="item in users" :key="item.Username" :label="renderUsername(item)" :value="item.Username">
                    </el-option>
                </el-select>
            </div>
        </div>
        <div :class="{'form-group':true, 'has-error': errors.has('share')}">
            <label class="col-sm-4 control-label">操作类型
            </label>
            <div class="col-sm-7 checkbox">
                <el-radio style="margin-left:-19px;margin-top:-5px;" size="small" v-model="form.share" :label="true">共享&nbsp;&nbsp;&nbsp;&nbsp;</el-radio>
                <el-radio style="margin-left:-19px;margin-top:-5px;" size="small" v-model="form.share" :label="false">取消共享</el-radio>
            </div>
        </div>
        <div :class="{'form-group':true, 'has-error': errors.has('subfetch')}">
            <label class="col-sm-4 control-label">其它选项
            </label>
            <div class="col-sm-7 checkbox">
                <el-checkbox style="margin-left:-19px;margin-top:-5px;" size="small" v-model.trim="form.subfetch" name="subfetch">
                    包含子节点&nbsp;&nbsp;
                </el-checkbox>
                <el-checkbox style="margin-left:-19px;margin-top:-5px;" size="small" v-model.trim="form.uptrace" name="uptrace">
                    包含父节点
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
        vtype: {
            type: String,
            default: ''
        }
    },
    data() {
        return {
            bLoading: false,
            cascades: [],
            users: [],
            form: this.defForm(),
        }
    },
    components: {
        FormDlg
    },
    computed: {
        title() {
            if(this.form.share) {
                return "共享节点";
            } else {
                return "取消共享节点";
            }
        }
    },
    methods: {
        defForm() {
            return {
                serial: '',
                code: '',
                name: '',
                target: 'cascade', // cascade, user
                cascade_id: '',
                username: '',
                subfetch: true,
                uptrace: false,
                devtree: true,
                share: true,
                share_all_channel: false,
            }
        },
        onHide() {
            this.form = this.defForm();
            if(this.vtype != '标准版') {
                this.form.target = 'cascade';
            } else {
                this.form.target = 'user';
            }
            this.bLoading = false;
            this.cascades = [];
            this.users = [];
            this.$emit("hide");
        },
        onShow() {
            this.errors.clear();
            this.$emit("show");
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
            if(this.form.target == 'cascade') {
                if(!this.form.cascade_id) {
                    this.$message({
                        type: 'error',
                        message: '请选择上级平台'
                    })
                    return;
                }
                if(this.form.share_all_channel && !this.form.share) {
                    this.$message({
                        type: 'error',
                        message: '上级平台已被标记为全部共享'
                    })
                    return;
                }
                $.post('/api/v1/cascade/sharenode', {
                    serial: this.form.serial,
                    code: this.form.code,
                    cascade_id: this.form.cascade_id,
                    share: this.form.share,
                    subfetch: this.form.subfetch,
                    uptrace: this.form.uptrace,
                    devtree: this.form.devtree,
                }).then(data => {
                    this.$refs['dlg'].hide();
                    this.$emit("submit");
                    this.$message({
                        type: 'success',
                        message: '操作成功'
                    })
                })
            } else if(this.form.target == 'user') {
                if(!this.form.username) {
                    this.$message({
                        type: 'error',
                        message: '请选择用户'
                    })
                    return;
                }
                if(this.form.share_all_channel && !this.form.share) {
                    this.$message({
                        type: 'error',
                        message: '用户已被标记为全部共享'
                    })
                    return;
                }
                $.post('/api/v1/user/sharenode', {
                    serial: this.form.serial,
                    code: this.form.code,
                    username: this.form.username,
                    share: this.form.share,
                    subfetch: this.form.subfetch,
                    uptrace: this.form.uptrace,
                    devtree: this.form.devtree,
                }).then(data => {
                    this.$refs['dlg'].hide();
                    this.$emit("submit");
                    this.$message({
                        type: 'success',
                        message: '操作成功'
                    })
                })
            }
        },
        show(data) {
            this.errors.clear();
            if(data) {
                this.form.serial = data.serial;
                this.form.code = data.code;
                this.form.name = data.customName || data.name || data.code;
                this.form.devtree = !!data.devtree;
            }
            if(this.vtype != '标准版') {
                this.form.target = 'cascade';
            } else {
                this.form.target = 'user';
            }
            this.$nextTick(() => {
                this.$refs['dlg'].show();
            })
        },
        loadCascade(q = '') {
            this.bLoading = true;
            $.get("/api/v1/cascade/list", {
                q: q,
                limit: 10,
            }).then(data => {
                this.cascades = data.CascadeList || [];
            }).always(() => {
                this.bLoading = false;
            })
        },
        loadUser(q = '') {
            this.bLoading = true;
            $.get("/api/v1/user/list", {
                q: q,
                limit: 10,
            }).then(data => {
                this.users = data.UserList || [];
            }).always(() => {
                this.bLoading = false;
            })
        },
        changeCascade(v) {
            if(!v) {
                this.form.cascade_id = '';
                this.form.share_all_channel = false;
            } else {
                this.form.cascade_id = v;
                for(var cas of this.cascades) {
                    if(cas.ID == v) {
                        this.form.share_all_channel = !!cas.ShareAllChannel;
                        break;
                    }
                }
            }
        },
        changeUser(v) {
            if(!v) {
                this.form.username = '';
                this.form.share_all_channel = false;
            } else {
                this.form.username = v;
                for(var user of this.users) {
                    if(user.Username == v) {
                        this.form.share_all_channel = !!user.HasAllChannel;
                        break;
                    }
                }
            }
        },
        renderUsername(item) {
            var name = item.Username || item.ID;
            if(item.HasAllChannel) {
                name += "(全部共享)"
            }
            return name;
        },
        renderCascadeName(item) {
            var name = item.Name||item.Serial||item.ID;
            if(item.ShareAllChannel) {
                name += "(全部共享)"
            }
            return name;
        }
    }
}
</script>
