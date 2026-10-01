<template>
    <FormDlg :title="title" @hide="onHide" @show="onShow" @submit="onSubmit" ref="dlg" :disabled="errors.any()">
        <div :class="{'form-group':true,'has-error': errors.has('parentid')}" v-show="form.parent_id || form.parent_name">
            <label for="input-parentid" class="col-sm-4 control-label">父节点
                <!-- <span class="text-red">*</span> -->
            </label>
            <div class="col-sm-7">
                <select class="form-control" id="input-parentid" v-model.trim="form.parent_id" data-vv-as="父节点" disabled="disabled">
                  <option :value="form.parent_id">{{form.parent_name}}</option>
                </select>
            </div>
        </div>
        <div :class="{'form-group':true,'has-error': errors.has('serial')}">
            <label for="input-serial" class="col-sm-4 control-label">设备编号
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" readonly="readonly" id="input-serial" name="serial" :value="form.serial" data-vv-as="设备编号" >
            </div>
        </div>
        <div :class="{'form-group':true,'has-error': errors.has('import_serial')}" v-if="bImport">
            <label for="input-serial" class="col-sm-4 control-label">导入设备
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <el-select filterable clearable remote :remote-method="loadDevice" :loading="bLoading" v-model="form.import_serial"
                    @change="changeDevice" @clear="clearDevice" @focus="!form.import_serial && loadDevice();" ref="selectDevice"
                    id="input-import-serial" name="import_serial" data-vv-as="导入设备" v-validate="'required'"
                    placeholder="可搜索" size="small" style="width:100%;">
                    <el-option v-for="item in devices" :key="item.ID" :label="item.CustomName||item.Name||item.ID" :value="item.ID">
                    </el-option>
                </el-select>
            </div>
        </div>
        <div :class="{'form-group':true,'has-error': errors.has('code')}" v-if="form.add || form.code">
            <label for="input-code" class="col-sm-4 control-label">{{ labelHeader }}编号
                <span class="text-red" v-if="form.add || form.custom">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" id="input-code" name="code" v-model.trim="form.code" :placeholder="codePlaceholder" data-vv-as="节点编号" v-validate="'required|regex:^[0-9]+$'" @keydown.enter="$el.querySelector('#input-name').focus()" v-if="form.add">
                <input type="text" class="form-control" id="input-code" name="code" v-model.trim="form.new_code" :placeholder="codePlaceholder" data-vv-as="节点编号" v-validate="'required|regex:^[0-9]+$'" @keydown.enter="$el.querySelector('#input-name').focus()" v-else-if="form.custom">
                <input type="text" class="form-control" readonly="readonly" id="input-code" name="code" v-model.trim="form.code" data-vv-as="节点编号" v-else>
            </div>
        </div>
        <div :class="{'form-group':true,'has-error': errors.has('custom_id')}" v-if="serverInfo.VersionType != '标准版' && !form.add && !form.custom && form.serial && form.code">
            <label for="input-custom-id" class="col-sm-4 control-label">自定义通道编号
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" id="input-custom-id" name="custom_id" placeholder="可选, 用于级联" v-model.trim="form.custom_id" data-vv-as="自定义通道编号" oninput="value=value.replace(/[^\d]/g,'')" @keydown.enter.prevent="$el.querySelector('#input-name').focus()">
            </div>
        </div>
        <div :class="{'form-group':true,'has-error': errors.has('name')}">
            <label for="input-name" class="col-sm-4 control-label">{{ labelHeader }}名称
            </label>
            <div class="col-sm-7">
                <div class="input-group" style="width: 100%;">
                    <input type="text" class="form-control" id="input-name" name="name" :placeholder="form.name||form.code||form.serial" v-model.trim="form.custom_name" data-vv-as="节点名称" @keydown.enter="onSubmit">
                    <span class="input-group-btn" v-show="form.custom_name || form.name">
                        <button type="button" class="btn btn-default" style="border-top-right-radius:0;border-bottom-right-radius:0;" :title="form.custom ? '拷贝名称' : form.custom_name ? '拷贝自定义名称' : '拷贝原名称'"
                            v-clipboard="form.custom_name || form.name" @success="$message({type:'success', message:'成功拷贝到粘贴板'})">
                            <i class="fa fa-copy"></i>
                        </button>
                    </span>
                </div>
            </div>
        </div>
    </FormDlg>
</template>

<script>
import FormDlg from 'components/FormDlg.vue'
import $ from 'jquery'

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
            bImport: false,
            bLoading: false,
            devices: [],
            form: this.defForm(),
        }
    },
    components: {
        FormDlg
    },
    computed: {
        title() {
            if(this.form.add) {
                return this.bImport ? "导入设备" : (this.form.custom ? "新建分组" : "新建节点");
            } else if(this.form.custom && this.form.code) {
                return "编辑分组";
            }
            return "编辑名称";
        },
        labelHeader() {
            if(this.form.add) {
                return "分组";
            } else if(this.form.custom && !this.form.code) {
                return "本域";
            } else if(this.form.custom) {
                return "分组";
            } else if(this.form.serial && !this.form.code) {
                return "设备";
            } else if(this.form.serial && this.form.code) {
                return "通道";
            }
            return "节点";
        },
        codePlaceholder() {
            if(!this.form.add) return "";
            if(!this.form.serial || this.form.serial.length < 4) return "1~20位数字";
            return `1~20位数字, 推荐使用行政编号, 如: ${this.form.serial.substring(0,4)}`;
        }
    },
    methods: {
        defForm() {
            return {
                id: '',
                serial: '',
                code: '',
                name: '',
                custom_id: '',
                custom_name: '',
                parent_id: '',
                parent_name: '根节点',
                custom: false,
                add: false,
                virtual: false,
                import_serial: '',
                new_code: '',
            }
        },
        onHide() {
            this.form = this.defForm();
            this.bLoading = false;
            this.bImport = false;
            this.devices = [];
            this.$emit("hide");
        },
        onShow() {
            this.errors.clear();
            this.$emit("show");
            if(this.title == "编辑名称") {
                this.$el.querySelector('#input-name').focus();
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
            if(this.form.virtual && this.form.add) {
                $.post('/api/v1/channel/virtual/add', {
                    serial: this.form.serial,
                    code: this.form.code,
                    name: this.form.custom_name,
                    parent_id: this.form.parent_id,
                }).then(data => {
                    this.$refs['dlg'].hide();
                    this.$emit("submit");
                })
            } else {
                $.post('/api/v1/channel/save', {
                    serial: this.form.serial,
                    code: this.form.code,
                    name: this.form.custom_name,
                    parent_id: this.form.parent_id,
                    add: this.form.add,
                    import_serial: this.form.import_serial,
                    new_code: this.form.new_code,
                    custom_id: this.form.custom_id,
                    custom: this.form.custom,
                }).then(data => {
                    this.$refs['dlg'].hide();
                    this.$emit("submit");
                })
            }
        },
        show(data) {
            this.errors.clear();
            if(data) {
                this.bImport = !!data.import;
                this.form.add = !!data.add;
                this.form.id = data.id;
                this.form.serial = data.serial;
                this.form.code = data.code;
                this.form.name = data.name;
                this.form.custom_id = data.customID;
                this.form.custom_name = data.customName;
                this.form.custom = !!data.custom;
                if(data.custom && data.code) {
                    this.form.new_code = data.code;
                    this.form.custom_name = data.name;
                    this.form.custom_id = '';
                    this.form.name = '';
                }
                if(data.parent) {
                    this.form.parent_id = data.parent.code;
                    this.form.parent_name = data.parent.customName || data.parent.name || data.parent.id;
                    this.form.virtual = data.parent.manufacturer === 'virtual';
                }
            }
            this.$nextTick(() => {
                this.$refs['dlg'].show();
            })
        },
        loadDevice(q = '') {
            this.bLoading = true;
            $.get("/api/v1/device/list", {
                q: q,
                limit: 10,
            }).then(data => {
                this.devices = data.DeviceList || [];
            }).always(() => {
                this.bLoading = false;
            })
        },
        changeDevice(v) {
            if(!v) {
                this.form.code = '';
                this.form.custom_name = '';
                this.form.import_serial = '';
            } else {
                this.form.code = v;
                if(v.length == 20 && v.substring(10, 13) != "216") {
                    this.form.code = v.substring(0, 10) + "216" + v.substring(13);
                }
                this.form.import_serial = v;
                this.$nextTick(() => {
                    this.form.custom_name = this.$refs['selectDevice'].selectedLabel;
                })
            }
        },
        clearDevice() {
            this.form.code = '';
            this.form.custom_name = '';
            this.form.import_serial = '';
        }
    }
}
</script>
