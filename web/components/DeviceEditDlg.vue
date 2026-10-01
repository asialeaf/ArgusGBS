<template>
    <FormDlg title="编辑设备" @hide="onHide" @show="onShow" @submit="onSubmit" ref="dlg" :disabled="errors.any()" size="modal-lgg">
        <div class="col-md-6">
            <div :class="{'form-group':true,'has-error': errors.has('serial')}">
                <label for="input-serial" class="col-sm-4 control-label">设备编号
                    <span class="text-red">*</span>
                </label>
                <div class="col-sm-7">
                    <input type="text" class="form-control" readonly="readonly" id="input-serial" name="serial" v-model.trim="form.serial" data-vv-as="设备编号" v-validate="'required'">
                </div>
            </div>
            <div :class="{'form-group':true,'has-error': errors.has('name')}">
                <label for="input-name" class="col-sm-4 control-label">自定义名称
                </label>
                <div class="col-sm-7">
                    <div class="input-group" style="width: 100%;">
                        <input type="text" class="form-control" id="input-name" name="name" :placeholder="form.name||'-'" v-model.trim="form.custom_name" data-vv-as="自定义名称" @keydown.enter="$el.querySelector('#input-password').focus()">
                        <span class="input-group-btn" v-show="form.custom_name || form.name">
                            <button type="button" class="btn btn-default" style="border-top-right-radius:0;border-bottom-right-radius:0;" :title="form.custom_name ? '拷贝自定义名称' : '拷贝原名称'"
                                v-clipboard="form.custom_name || form.name" @success="$message({type:'success', message:'成功拷贝到粘贴板'})">
                                <i class="fa fa-copy"></i>
                            </button>
                        </span>
                    </div>
                </div>
            </div>
            <div :class="{'form-group':true,'has-error': errors.has('password')}" v-show="useSeparateDevicePassword">
                <label for="input-password" class="col-sm-4 control-label">接入密码
                </label>
                <div class="col-sm-7">
                    <input type="text" class="form-control" id="input-password" name="password" v-model.trim="form.password" placeholder="默认使用 livecms.ini > sip > device_password" data-vv-as="接入密码" @keydown.enter="$el.querySelector('#input-contact-ip').focus()">
                </div>
            </div>
            <div :class="{'form-group':true,'has-error': errors.has('contact_ip')}">
                <label for="input-contact-ip" class="col-sm-4 control-label">信令联络 IP
                </label>
                <div class="col-sm-7">
                    <input type="text" class="form-control" id="input-contact-ip" name="contact_ip" v-model.trim="form.contact_ip" placeholder="默认使用 livecms.ini > sip > host" data-vv-as="信令联络 IP" @keydown.enter="$el.querySelector('#input-recv-stream-ip').focus()">
                </div>
            </div>
            <div :class="{'form-group':true,'has-error': errors.has('sms_id')}">
                <label for="input-sms-id" class="col-sm-4 control-label">SMS ID
                </label>
                <div class="col-sm-7">
                    <select class="form-control" id="input-sms-id" name="sms_id" v-model.trim="form.sms_id" data-vv-as="SMS" v-validate="">
                        <option value="">自动选择负载最小</option>
                        <option v-for="(sms, idx) in smsList" :value="sms.id" :key="idx">{{sms.name}}</option>
                    </select>
                </div>
            </div>
            <!-- <div :class="{'form-group':true,'has-error': errors.has('sms_group_id')}" v-show="smsGroupList.length > 0 && !form.sms_id"> -->
            <div :class="{'form-group':true,'has-error': errors.has('sms_group_id')}">
                <label for="input-sms-group-id" class="col-sm-4 control-label">SMS 分组
                </label>
                <div class="col-sm-7">
                    <select class="form-control" id="input-sms-group-id" name="sms_group_id" v-model.trim="form.sms_group_id" data-vv-as="SMSGroup" v-validate="">
                        <option value="">自动选择负载最小</option>
                        <option v-for="(group, idx) in smsGroupList" :value="group.id" :key="idx">{{group.name}}</option>
                    </select>
                </div>
            </div>
            <div :class="{'form-group':true, 'has-error': errors.has('recv_stream_ip')}">
                <label for="input-recv-stream-ip" class="col-sm-4 control-label">收流 IP
                </label>
                <div class="col-sm-7">
                    <input type="text" class="form-control" id="input-recv-stream-ip" name="recv_stream_ip" v-model.trim="form.recv_stream_ip" placeholder="默认使用 livesms.ini > sip > host" data-vv-as="收流 IP" @keydown.enter="$el.querySelector('#input-catalog-interval').focus()">
                </div>
            </div>
            <div :class="{'form-group':true, 'has-error': errors.has('drop_channel_type')}">
                <label for="input-drop-channel-type" class="col-sm-4 control-label">过滤通道类型</label>
                <div class="col-sm-7">
                    <el-select id="input-drop-channel-type" style="width:100%;" size="medium" v-model.trim="dropChannelTypes" multiple filterable allow-create default-first-option placeholder="选填, 需要丢弃的通道类型/编号">
                        <el-option v-for="(item, idx) in innerChannelTypes" :key="idx" :label="`${idx} - ${item}`" :value="idx">
                        </el-option>
                    </el-select>
                </div>
            </div>
            <div :class="{'form-group': true,'has-error': errors.has('longitude')}" v-show="debug || showPosition">
                <label for="input-position-longitude" class="col-sm-4 control-label">经度
                </label>
                <div class="col-sm-7">
                    <input type="text" class="form-control" id="input-position-longitude" name="longitude" v-validate="'decimal'"
                        @paste="onPasteLngLat"
                        v-model.trim="form.longitude" data-vv-as="经度">
                </div>
            </div>
            <div :class="{'form-group': true,'has-error': errors.has('latitude')}" v-show="debug || showPosition">
                <label for="input-position-latitude" class="col-sm-4 control-label">纬度
                </label>
                <div class="col-sm-7">
                    <input type="text" class="form-control" id="input-position-latitude" name="latitude" v-validate="'decimal'"
                        @paste="onPasteLngLat"
                        v-model.trim="form.latitude" data-vv-as="纬度">
                </div>
            </div>
        </div>
        <div class="col-md-6">
            <div :class="{'form-group':true, 'has-error': errors.has('catalog_interval')}">
                <label for="input-catalog-interval" class="col-sm-4 control-label">通道更新周期(秒)
                </label>
                <div class="col-sm-7">
                    <input type="text" class="form-control" id="input-catalog-interval" name="catalog_interval" v-model.trim="form.catalog_interval" v-validate="'numeric'" placeholder="3600" @keydown.enter="$el.querySelector('#input-subscribe-interval').focus()">
                </div>
            </div>
            <div :class="{'form-group':true, 'has-error': errors.has('subscribe_interval')}" v-show="debug || showSubscribeInterval">
                <label for="input-subscribe-interval" class="col-sm-4 control-label">订阅周期(秒)
                </label>
                <div class="col-sm-7">
                    <input type="text" class="form-control" id="input-subscribe-interval" name="subscribe_interval" v-model.trim="form.subscribe_interval" v-validate="'integer'" placeholder="默认使用全局订阅配置, 负数不订阅" @keydown.enter="onSubmit">
                </div>
            </div>
            <div :class="{'form-group': true, 'has-error': errors.has('charset')}">
                <label for="input-charset" class="col-sm-4 control-label">字符集
                </label>
                <div class="col-sm-7">
                    <select class="form-control" id="input-charset" name="charset" v-model.trim="form.charset" data-vv-as="字符集" v-validate="">
                        <option value="">自动识别</option>
                        <option value="GB2312">GB2312</option>
                        <option value="GB18030">GB18030</option>
                        <option value="UTF-8">UTF-8</option>
                    </select>
                </div>
            </div>
            <div :class="{'form-group': true, 'has-error': errors.has('stream_mode')}">
                <label for="input-stream-mode" class="col-sm-4 control-label">码流索引
                </label>
                <div class="col-sm-7">
                    <el-select id="input-stream-mode" style="width:100%;" size="medium" v-model.trim="form.stream_mode"
                        clearable filterable allow-create default-first-option placeholder="自动选择">
                        <el-option v-for="(item, idx) in innerStreamModes" :key="idx" :label="idx ? `${idx} - ${item}` : `${item}`" :value="idx">
                        </el-option>
                    </el-select>
                </div>
            </div>
            <div :class="{'form-group':true,'has-error': errors.has('media_transport')}">
                <label class="col-sm-4 control-label">流传输模式
                </label>
                <div class="col-sm-7">
                    <el-radio-group v-model.trim="form.media_transport" size="mini" @change="onMediaTransportChange">
                        <el-radio-button label="TCP"></el-radio-button>
                        <el-radio-button label="UDP"></el-radio-button>
                    </el-radio-group>
                    &nbsp;&nbsp;
                    <el-radio-group v-model.trim="form.media_transport_mode" size="mini" v-show="form.media_transport == 'TCP'">
                        <el-radio-button label="passive">被动</el-radio-button>
                        <el-radio-button label="active">主动</el-radio-button>
                    </el-radio-group>
                </div>
            </div>
            <div :class="{'form-group':true,'has-error': errors.has('subscribe_items')}">
                <label class="col-sm-4 control-label">订阅项目
                </label>
                <div class="col-sm-7 checkbox">
                    <el-checkbox style="margin-left:-19px;margin-top:-5px;" size="small" v-model.trim="form.catalog_subscribe" name="catalog_subscribe" :disabled="form.subscribe_interval < 0">
                        目录&nbsp;&nbsp;
                    </el-checkbox>
                    <el-checkbox style="margin-left:-19px;margin-top:-5px;" size="small" v-model.trim="form.alarm_subscribe" name="alarm_subscribe" :disabled="form.subscribe_interval < 0">
                        报警&nbsp;&nbsp;
                    </el-checkbox>
                    <el-checkbox style="margin-left:-19px;margin-top:-5px;" size="small" v-model.trim="form.position_subscribe" name="position_subscribe" :disabled="form.subscribe_interval < 0">
                        位置&nbsp;&nbsp;
                    </el-checkbox>
                    <el-checkbox style="margin-left:-19px;margin-top:-5px;" size="small" v-model.trim="form.ptz_subscribe" name="ptz_subscribe" :disabled="form.subscribe_interval < 0">
                        PTZ(2022)
                    </el-checkbox>
                </div>
            </div>
            <div :class="{'form-group':true,'has-error': errors.has('record')}">
                <label class="col-sm-4 control-label">设备录像
                </label>
                <div class="col-sm-7 checkbox">
                    <el-checkbox style="margin-left:-19px;margin-top:-5px;" size="small" v-model.trim="form.record_center" name="record_center">
                        中心检索&nbsp;&nbsp;
                    </el-checkbox>
                    <el-checkbox style="margin-left:-19px;margin-top:-5px;" size="small" v-model.trim="form.record_indistinct" name="record_indistinct">
                        模糊查询
                    </el-checkbox>
                </div>
            </div>
            <div :class="{'form-group':true,'has-error': errors.has('catalog')}">
                <label class="col-sm-4 control-label">目录结构
                </label>
                <div class="col-sm-7 checkbox">
                    <el-checkbox style="margin-left:-19px;margin-top:-5px;" size="small" v-model.trim="form.civil_code_first" name="civil_code_first">
                        行政区域优先&nbsp;&nbsp;
                    </el-checkbox>
                    <el-checkbox style="margin-left:-19px;margin-top:-5px;" size="small" v-model.trim="form.keep_original_tree" name="keep_original_tree" v-show="debug">
                        保持原样
                    </el-checkbox>
                </div>
            </div>
        </div>
        <div class="clearfix"></div>

        <template slot="footer">
            <div class="inline text-orange" v-if="alertMsg">
                <small>
                    <strong><i class="fa fa-info-circle"></i> 提示 : &nbsp;&nbsp;</strong>
                    {{alertMsg}}
                    &nbsp;&nbsp;
                </small>
            </div>
        </template>
    </FormDlg>
</template>

<script>
import FormDlg from 'components/FormDlg.vue';
import $ from 'jquery';

export default {
    props: {
        useSeparateDevicePassword: {
            type: Boolean,
            default: false
        },
    },
    data() {
        return {
            form: this.defForm(),
            smsList: [],
            smsGroupList: [],
            dropChannelTypes: [],
            innerChannelTypes: {
                "134": "报警输入",
                "135": "报警输出",
                "136": "语音输入",
                "137": "语音输出",
                "200": "中心信令",
                "215": "业务分组",
                "216": "虚拟组织",
            },
            innerStreamModes: {
                "": "自动选择",
                "stream:0": "主码流",
                "stream:1": "子码流",
                "streamnumber:0": "主码流(2022)",
                "streamnumber:1": "子码流(2022)",
                "streamprofile:0": "主码流",
                "streamprofile:1": "子码流",
                "streamMode:MAIN": "主码流",
                "streamMode:SUB": "子码流",
            },
            debug: false,
            showSubscribeInterval: false,
            showPosition: false,
        }
    },
    components: {
        FormDlg
    },
    beforeDestroy() {
        $(this.$el).off("keydown", this.keyDown);
    },
    computed: {
        alertMsg() {
            var list = [];
            if(this.form.stream_mode) {
                list.push("码流索引, 需要下级支持");
            }
            var msg = "";
            if(list.length > 0) {
                msg = list.join(", ")
            }
            return msg;
        }
    },
    methods: {
        defForm() {
            return {
                serial: '',
                name: '',
                custom_name: '',
                media_transport: 'UDP',
                media_transport_mode: 'passive',
                stream_mode: '',
                sms_id: '',
                sms_group_id: '',
                recv_stream_ip: '',
                contact_ip: '',
                charset: '',
                catalog_interval: 3600,
                subscribe_interval: 0,
                catalog_subscribe: false,
                alarm_subscribe: false,
                position_subscribe: false,
                ptz_subscribe: false,
                record_center: false,
                record_indistinct: false,
                civil_code_first: false,
                keep_original_tree: false,
                password: '',
                drop_channel_type: '',
                longitude: 0,
                latitude: 0,
            }
        },
        onHide() {
            this.debug = false;
            this.showSubscribeInterval = false;
            this.showPosition = false;
            this.form = this.defForm();
            this.dropChannelTypes = [];
            this.$emit("hide");
            $(this.$el).off("keydown", this.keyDown);
        },
        onShow() {
            this.errors.clear();
            this.$emit("show");
            $(this.$el).on("keydown", this.keyDown);
        },
        onPasteLngLat(e) {
            try {
                var vals = e.clipboardData.getData('Text').trim().split(/\s*,\s*/);
                if(vals.length == 2) {
                    this.form.longitude = vals[0];
                    this.form.latitude = vals[1];
                    e.preventDefault();
                }
            } catch (error) {
                console.log(error);
            }
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
        onMediaTransportChange(label) {
            if("UDP" == label) {
                this.form.media_transport_mode = "passive";
            }
        },
        fetchSMSList() {
            return new Promise((resolve, reject) => {
                $.get("/api/v1/sms/list").then(ret => {
                    resolve(ret.map(sms => ({
                        id: sms.Serial||"",
                        name: sms.Serial||"",
                        group: sms.GroupID||"",
                    })));
                }).fail(() => {
                    resolve([]);
                })
            })
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
            // if(this.form.subscribe_interval > 0) {
            //     if(!this.form.catalog_subscribe && !this.form.alarm_subscribe && !this.form.position_subscribe) {
            //         this.$message({
            //             type: 'error',
            //             message: '请勾选订阅项目',
            //         })
            //         return;
            //     }
            // }
            var data = Object.assign({}, this.form);
            data["drop_channel_type"] = this.dropChannelTypes.join(",");
            delete data["name"];
            if(!data['longitude']) data['longitude'] = 0;
            if(!data['latitude']) data['latitude'] = 0;
            $.post('/api/v1/device/setinfo', data).then(data => {
                this.$refs['dlg'].hide();
                this.$emit("submit");
            })
        },
        async show(data) {
            this.errors.clear();
            if(data) {
                Object.assign(this.form, data);
            }
            if(this.form["drop_channel_type"]) {
                this.dropChannelTypes = this.form["drop_channel_type"].split(",");
            }
            if(this.form['subscribe_interval'] < 0) {
                this.showSubscribeInterval = true;
            }
            if(this.form['longitude'] || this.form['latitude']) {
                this.showPosition = true;
            }
            var smsGroupList = [];
            var smsList = await this.fetchSMSList();
            for (var sms of smsList) {
                if(sms.group && !smsGroupList.some(group => (group.id == sms.group))) {
                    smsGroupList.push({
                        id: sms.group,
                        name: sms.group,
                    });
                }
            }
            var smsid = this.form['sms_id'];
            if(smsid && !smsList.some(sms => (sms.id == smsid))) {
                smsList.push({
                    id: smsid,
                    name: `${smsid}(not found)`,
                });
            }
            var smsGroupID = this.form['sms_group_id'];
            if(smsGroupID && !smsGroupList.some(group => (group.id == smsGroupID))) {
                smsGroupList.push({
                    id: smsGroupID,
                    name: `${smsGroupID}(not found)`,
                });
            }
            this.smsList = smsList;
            this.smsGroupList = smsGroupList;
            if(!this.form['subscribe_interval']) {
                this.form['subscribe_interval'] = '';
            }
            if(!this.form['longitude']) {
                this.form['longitude'] = '';
            }
            if(!this.form['latitude']) {
                this.form['latitude'] = '';
            }
            this.$nextTick(() => {
                this.$refs['dlg'].show();
            })
        }
    }
}
</script>
