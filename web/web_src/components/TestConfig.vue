<template>
<div>
    <div class="input-area form-horizontal">
        <div class="form-group">
            <label class="col-sm-4 control-label">配置参数类型
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <select class="form-control" name="config_type" v-model.trim="config_type">
                    <option value="BasicParam">基本参数</option>
                    <option value="OSDConfig">OSD参数</option>
                    <option value="SnapShotConfig">图像抓拍配置</option>
                    <option value="FrameMirror">画面翻转配置</option>
                    <option value="VideoParamAttribute">视频参数属性</option>
                    <option value="VideoParamOpt">视频参数范围</option>
                    <option value="VideoRecordPlan">录像计划配置</option>
                    <option value="VideoAlarmRecord">报警录像配置</option>
                    <option value="PictureMask">视频画面遮挡配置</option>
                    <option value="AlarmReport">报警上报开关配置</option>
                    <option value="SVACEncodeConfig">SVAC 编码</option>
                    <option value="SVACDecodeConfig">SVAC 解码</option>
                </select>
            </div>
        </div>
    </div>
    <el-tabs tab-position="left" v-model="tab_label">
         <el-tab-pane label="基本参数" name="BasicParam">
            <div class="basic-param-area form-horizontal">
                <div class="form-group">
                    <div class="col-sm-offset-4 col-sm-7">
                        <button role="button" class="btn btn-info" @click.prevent="query" :disabled="loading">配置获取<span v-show="loading">...</span></button>
                        <button role="button" class="btn btn-primary" @click.prevent="sendBasic" :disabled="sending">发送<span v-show="sending">...</span></button>
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-4 control-label">设备名称
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="text" class="form-control" v-model="BasicParam.Name">
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-4 control-label">注册过期
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="text" class="form-control" v-model="BasicParam.Expiration">
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-4 control-label">心跳周期
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="text" class="form-control" v-model="BasicParam.HeartBeatInterval">
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-4 control-label">超时次数
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="text" class="form-control" v-model="BasicParam.HeartBeatCount">
                    </div>
                </div>
            </div>
        </el-tab-pane>
        <el-tab-pane label="OSD参数" name="OSDConfig">
            <div class="osd-config-area form-horizontal">
                <div class="form-group">
                    <div class="col-sm-offset-4 col-sm-7">
                        <button role="button" class="btn btn-info" @click.prevent="query" :disabled="loading">配置获取<span v-show="loading">...</span></button>
                        <button role="button" class="btn btn-primary" @click.prevent="sendOSD" :disabled="sending">发送<span v-show="sending">...</span></button>
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-4 control-label">窗口长度像素值
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="number" min="0" class="form-control" v-model="OSDConfig.Length">
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-4 control-label">窗口宽度像素值
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="number" min="0" class="form-control" v-model="OSDConfig.Width">
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-4 control-label">时间X坐标
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="number" class="form-control" v-model="OSDConfig.TimeX">
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-4 control-label">时间Y坐标
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="number" class="form-control" v-model="OSDConfig.TimeY">
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-4 control-label">开启时间
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <el-radio-group size="mini" v-model="OSDConfig.TimeEnable">
                            <el-radio :label="1">开启</el-radio>
                            <el-radio :label="0">关闭</el-radio>
                        </el-radio-group>
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-4 control-label">时间类型
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <select class="form-control" v-model.trim="OSDConfig.TimeType">
                            <option value=""></option>
                            <option value="0">YYYY-MM-DD HH:MM:SS</option>
                            <option value="1">YYYY年MM月DD日HH:MM:SS</option>
                        </select>
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-4 control-label">开启文字
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <el-radio-group size="mini" v-model="OSDConfig.TextEnable">
                            <el-radio :label="1">开启</el-radio>
                            <el-radio :label="0">关闭</el-radio>
                        </el-radio-group>
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-4 control-label">内容设置
                    </label>
                    <div class="col-sm-7">
                        <el-input type="textarea" :autosize="{minRows:5, maxRows:8}" placeholder="每行1条，最多8条，每行内部用/分隔文本和坐标，格式为：文字/X坐标/Y坐标 例如：大门口/450/300" v-model="OSDConfig.TextContent"></el-input>
                    </div>
                </div>
            </div>
        </el-tab-pane>
        <el-tab-pane label="图像抓拍" name="SnapShotConfig">
            <div class="snapshot-config-area form-horizontal">
                <div class="form-group">
                    <div class="col-sm-offset-4 col-sm-7">
                        <button role="button" class="btn btn-info" @click.prevent="query" :disabled="loading">配置获取<span v-show="loading">...</span></button>
                        <button role="button" class="btn btn-primary" @click.prevent="sendSnapShot" :disabled="sending">发送<span v-show="sending">...</span></button>
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-4 control-label">连拍张数
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="number" min="0" v-model.number="SnapShot.SnapNum" class="form-control" style="width: 100px;">
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-4 control-label">抓拍间隔(秒)
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="number" min="0" v-model.number="SnapShot.Interval" class="form-control" style="width: 100px;">
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-4 control-label">上传地址
                    </label>
                    <div class="col-sm-7">
                        <input type="text" class="form-control" v-model="SnapShot.UploadURL">
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-4 control-label">SessionID
                    </label>
                    <div class="col-sm-7">
                        <input type="text" class="form-control" v-model="SnapShot.SessionID">
                    </div>
                </div>
            </div>
            <div class="clearfix" v-if="SnapFiles.length > 0">
                <p>快照预览</p>
                <el-popover :open-delay="500" :close-delay="10" placement="left" :title="`快照预览[${idx + 1}]`" width="500" trigger="hover" v-for="(url, idx) in SnapFiles" :key="idx">
                    <img :src="url" :alt="url" style="width:100%;height:100%;">
                    <img onerror="this.src='/images/default_snap.png';" :src="url" :alt="url" width="160" height="90" slot="reference" style="margin:2px;">
                </el-popover>
            </div>
        </el-tab-pane>
        <el-tab-pane label="画面翻转" name="FrameMirror">
            <div class="mirror-config-area form-horizontal">
                <div class="form-group">
                    <div class="col-sm-offset-4 col-sm-7">
                        <button role="button" class="btn btn-info" @click.prevent="query" :disabled="loading">配置获取<span v-show="loading">...</span></button>
                        <button role="button" class="btn btn-primary" @click.prevent="sendMirror" :disabled="sending">发送<span v-show="sending">...</span></button>
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-4 control-label">画面翻转
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <el-radio-group size="mini" v-model="FrameMirror">
                            <el-radio :label="1">左右翻转</el-radio>
                            <el-radio :label="2">上下翻转</el-radio>
                            <el-radio :label="3">中心翻转</el-radio>
                            <el-radio :label="0">关闭</el-radio>
                        </el-radio-group>
                    </div>
                </div>
            </div>
        </el-tab-pane>
        <el-tab-pane label="视频参数属性" name="VideoParamAttribute">
            <div class="video-param-attribute-area form-horizontal">
                <div class="form-group">
                    <div class="col-sm-offset-4 col-sm-7">
                        <button role="button" class="btn btn-info" @click.prevent="query" :disabled="loading">配置获取<span v-show="loading">...</span></button>
                        <button role="button" class="btn btn-primary" @click.prevent="sendVideoParamAttr" :disabled="sending">发送<span v-show="sending">...</span></button>
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-4 control-label">视频流编号
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <select class="form-control" v-model.number="streamNumber">
                            <option v-for="(item, idx) in VideoParamAttribute.Item" :key="idx" :value="idx">{{formatStreamNumber(idx)}}</option>
                        </select>
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-4 control-label">编码格式
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="text" v-model="VideoParamAttribute.Item[streamNumber].VideoFormat" class="form-control">
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-4 control-label">分辨率
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="text" v-model="VideoParamAttribute.Item[streamNumber].Resolution" class="form-control">
                    </div>
                </div>
               <div class="form-group">
                    <label class="col-sm-4 control-label">帧率
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="text" v-model="VideoParamAttribute.Item[streamNumber].FrameRate" class="form-control">
                    </div>
                </div>
               <div class="form-group">
                    <label class="col-sm-4 control-label">码率类型
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="text" v-model="VideoParamAttribute.Item[streamNumber].BitRateType" class="form-control">
                    </div>
                </div>
               <div class="form-group">
                    <label class="col-sm-4 control-label">码率大小
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="text" v-model="VideoParamAttribute.Item[streamNumber].VideoBitRate" class="form-control">
                    </div>
                </div>
            </div>
        </el-tab-pane>
        <el-tab-pane label="其他参数" name="OtherConfig">
            <div class="other-config-area form-horizontal">
                <div class="form-group">
                    <div class="col-sm-offset-4 col-sm-7">
                        <button role="button" class="btn btn-info" @click.prevent="query" :disabled="loading">配置获取<span v-show="loading">...</span></button>
                        <button role="button" class="btn btn-primary" @click.prevent="sendOther" :disabled="sending">发送<span v-show="sending">...</span></button>
                    </div>
                </div>
                <div class="form-group">
                    <div class="col-sm-offset-2 col-sm-9">
                        <el-input type="textarea" :autosize="{minRows:10, maxRows:30}" placeholder="参数内容" v-model="OtherConfig"></el-input>
                    </div>
                </div>
            </div>
        </el-tab-pane>
    </el-tabs>
    <div class="play-area" v-if="videoUrl">
        <div class="row">
            <div class="col-sm-10 col-sm-offset-1 col-lg-8 col-lg-offset-2">
                <LivePlayer ref="player" :videoUrl="videoUrl" live stretch :muted="!audio" :controls="false">
                </LivePlayer>
            </div>
        </div>
    </div>
</div>
</template>

<script>
import { mapState } from "vuex";
import LivePlayer from "@liveqing/liveplayer";

export default {
    data() {
        return {
            loading: false,
            sending: false,
            config_type: "BasicParam",
            tab_label: "BasicParam",
            BasicParam: {
                Name: "",
                Expiration: "",
                HeartBeatInterval: "",
                HeartBeatCount: "",
            },
            OSDConfig: {
                Length: "",
                Width: "",
                TimeX: "",
                TimeY: "",
                TimeEnable: "",
                TimeType: "",
                TextEnable: "",
                TextContent: "",
                SumNum: 0,
                Item: [],
            },
            SnapShot: {
                SnapNum: "",
                Interval: "",
                UploadURL: "",
                SessionID: "",
            },
            FrameMirror: "",
            VideoParamAttribute: {
                Item: [{
                    StreamNumber: "0",
                    VideoFormat: "",
                    Resolution: "",
                    FrameRate: "",
                    BitRateType: "",
                    VideoBitRate: "",
                }],
            },
            SnapFiles: [],
            OtherConfig: "",
            regConfig: null,
            videoUrl: "",
            audio: false,
            streamNumber: 0,
        }
    },
    computed: {
        ...mapState(["msg"]),
    },
    watch: {
        tab_label: function(newVal, oldVal) {
            this.OtherConfig = "";
            this.regConfig = null;
            this.SnapFiles = [];
            this.videoUrl = "";
            this.$store.commit("updateResult", null);
            if(["BasicParam", "OSDConfig", "SnapShotConfig", "FrameMirror", "VideoParamAttribute"].indexOf(newVal) >= 0) {
                this.config_type = newVal;
            } else if(["BasicParam", "OSDConfig", "SnapShotConfig", "FrameMirror", "VideoParamAttribute"].indexOf(this.config_type) >= 0) {
                this.config_type = "VideoParamOpt";
            }
        },
        config_type: function(newVal, oldVal) {
            this.OtherConfig = "";
            this.regConfig = null;
            this.SnapFiles = [];
            this.videoUrl = "";
            this.$store.commit("updateResult", null);
            if(["BasicParam", "OSDConfig", "SnapShotConfig", "FrameMirror", "VideoParamAttribute"].indexOf(newVal) >= 0) {
                this.tab_label = newVal;
            } else {
                this.tab_label = "OtherConfig";
            }
        },
        msg: function(newVal, oldVal) {
            if(!this.regConfig) return;
            if(!newVal || newVal.length <= oldVal.length) return;
            var last = newVal.substr(oldVal.length);
            if(last.startsWith("<<<<<<")) {
                var match = this.regConfig.exec(last);
                if(match && match.length > 0) {
                    this.OtherConfig = match[0];
                    this.regConfig = null;
                }
            }
        },
    },
    components: { LivePlayer },
    methods: {
        async query() {
            if(this.tab_label == "OSDConfig" || this.tab_label == "FrameMirror" || this.tab_label == "VideoParamAttribute") {
                var streamInfo = await this.$store.dispatch("play");
                if(streamInfo) {
                    var videoUrl = streamInfo.HLS || "";
                    if(this.flvSupported()) {
                        videoUrl = this.isIE() ? streamInfo.WS_FLV || "" : streamInfo.FLV || "";
                    }
                    this.audio = !!streamInfo.AudioEnable;
                    this.videoUrl = videoUrl;
                }
            } else if(this.tab_label == "SnapShotConfig") {
                this.SnapFiles = [];
            } else if(this.tab_label == "OtherConfig") {
                var tag = this.config_type;
                if(this.config_type == "SnapShotConfig") {
                    tag = "SnapShot";
                }
                this.regConfig = new RegExp(`<${tag}[\\s\\S]+<\/${tag}>`, "m");
            }

            const select = this.$el.querySelector('select[name="config_type"]');
            let label = this.config_type;
            if (select) {
                const option = Array.from(select.options).find(opt => opt.value === this.config_type);
                if (option) label = option.text;
            }
            if(!label.endsWith("配置")) {
                label += "配置";
            }
            this.loading = true;
            await this.$store.dispatch("connect", {
                topic: `获取${label}`,
            });

            $.get("/api/v1/device/fetchconfig", {
                serial: this.$store.state.serial,
                code: this.$store.state.code||this.$store.state.serial,
                config_type: this.config_type,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "配置获取成功"
                })
                this.$store.commit("updateResult", ret);
                if(this.config_type == "BasicParam" && ret.BasicParam) {
                    Object.assign(this.BasicParam, ret.BasicParam);
                } else if(this.config_type == "OSDConfig" && ret.OSDConfig) {
                    Object.assign(this.OSDConfig, ret.OSDConfig);
                    this.OSDConfig.TextContent = "";
                    if(ret.OSDConfig.SumNum > 0) {
                        var content = "";
                        for(var item of ret.OSDConfig.Item) {
                            if(content) {
                                content += "\r\n";
                            }
                            content += `${item.Text}/${item.X}/${item.Y}`;
                        }
                        this.OSDConfig.TextContent = content;
                    }
                } else if(this.config_type == "SnapShotConfig" && ret.SnapShot) {
                    Object.assign(this.SnapShot, ret.SnapShot);
                } else if(this.config_type == "FrameMirror") {
                    this.FrameMirror = ret.FrameMirror;
                } else if(this.config_type == "VideoParamAttribute" && ret.VideoParamAttribute && ret.VideoParamAttribute.Item && ret.VideoParamAttribute.Item.length > 0) {
                    if(this.streamNumber >= ret.VideoParamAttribute.Item.length) {
                        this.streamNumber = 0;
                    }
                    Object.assign(this.VideoParamAttribute, ret.VideoParamAttribute);
                }
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).fail(xhr => {
                this.$store.commit("updateResult", null);
            }).always(() => {
                this.loading = false;
            })
        },
        async sendBasic() {
            this.sending = true;
            await this.$store.dispatch("connect", {
                topic: "下发基本参数配置",
            });
            const xml = `<BasicParam>
<Name>${this.BasicParam.Name}</Name>
<Expiration>${this.BasicParam.Expiration}</Expiration>
<HeartBeatCount>${this.BasicParam.HeartBeatCount}</HeartBeatCount>
<HeartBeatInterval>${this.BasicParam.HeartBeatInterval}</HeartBeatInterval>
</BasicParam>`;
            $.post("/api/v1/control/deviceconfig", {
                serial: this.$store.state.serial,
                code: this.$store.state.code||this.$store.state.serial,
                BasicParam: xml,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "基本参数配置成功"
                })
                this.$store.commit("updateResult", ret);
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).fail(xhr => {
                this.$store.commit("updateResult", null);
            }).always(() => {
                this.sending = false;
            })
        },
        async sendOSD() {
            var streamInfo = await this.$store.dispatch("play");
            if(streamInfo) {
                var videoUrl = streamInfo.HLS || "";
                if(this.flvSupported()) {
                    videoUrl = this.isIE() ? streamInfo.WS_FLV || "" : streamInfo.FLV || "";
                }
                this.audio = !!streamInfo.AudioEnable;
                this.videoUrl = videoUrl;
            }

            this.sending = true;
            await this.$store.dispatch("connect", {
                topic: "下发OSD参数配置",
            });
            var sumNum = 0;
            var items = "";
            if(this.OSDConfig.TextContent) {
                var lines = this.OSDConfig.TextContent.split(/\r?\n/);
                for(var line of lines) {
                    var vals = line.split('/');
                    if(vals.length == 3) {
                        items += `<Item><Text>${vals[0]}</Text><X>${parseInt(vals[1])||0}</X><Y>${parseInt(vals[2])||0}</Y></Item>\r\n`;
                        sumNum++;
                    }
                }
            }
            const xml = `<OSDConfig>
<Length>${this.OSDConfig.Length}</Length>
<Width>${this.OSDConfig.Width}</Width>
<TimeX>${this.OSDConfig.TimeX}</TimeX>
<TimeY>${this.OSDConfig.TimeY}</TimeY>
<TimeEnable>${this.OSDConfig.TimeEnable}</TimeEnable>
<TimeType>${this.OSDConfig.TimeType}</TimeType>
<TextEnable>${this.OSDConfig.TextEnable}</TextEnable>
<SumNum>${sumNum}</SumNum>
${items}</OSDConfig>`;
            $.post("/api/v1/control/deviceconfig", {
                serial: this.$store.state.serial,
                code: this.$store.state.code,
                OSDConfig: xml,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "OSD参数配置成功"
                })
                this.$store.commit("updateResult", ret);
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).fail(xhr => {
                this.$store.commit("updateResult", null);
            }).always(() => {
                this.sending = false;
            })
        },
        async sendSnapShot() {
            this.SnapFiles = [];
            this.sending = true;
            await this.$store.dispatch("connect", {
                topic: "下发图像抓拍配置",
            });

            $.post("/api/v1/control/snapshotconfig", {
                serial: this.$store.state.serial,
                code: this.$store.state.code,
                snap_num: this.SnapShot.SnapNum,
                interval: this.SnapShot.Interval,
                upload_url: this.SnapShot.UploadURL,
                session_id: this.SnapShot.SessionID,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "图像抓拍配置成功"
                })
                this.SnapFiles = ret.DownloadList||[];
                this.$store.commit("updateResult", ret);
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).fail(xhr => {
                this.$store.commit("updateResult", null);
            }).always(() => {
                this.sending = false;
            })
        },
        async sendMirror() {
            var streamInfo = await this.$store.dispatch("play");
            if(streamInfo) {
                var videoUrl = streamInfo.HLS || "";
                if(this.flvSupported()) {
                    videoUrl = this.isIE() ? streamInfo.WS_FLV || "" : streamInfo.FLV || "";
                }
                this.audio = !!streamInfo.AudioEnable;
                this.videoUrl = videoUrl;
            }

            this.sending = true;
            await this.$store.dispatch("connect", {
                topic: "下发画面翻转配置",
            });
            const xml = `<FrameMirror>${this.FrameMirror ? this.FrameMirror : '0'}</FrameMirror>`;
            $.post("/api/v1/control/deviceconfig", {
                serial: this.$store.state.serial,
                code: this.$store.state.code,
                FrameMirror: xml,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "画面翻转配置成功"
                })
                this.$store.commit("updateResult", ret);
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).fail(xhr => {
                this.$store.commit("updateResult", null);
            }).always(() => {
                this.sending = false;
            })
        },
        async sendVideoParamAttr() {
            var streamInfo = await this.$store.dispatch("play");
            if(streamInfo) {
                var videoUrl = streamInfo.HLS || "";
                if(this.flvSupported()) {
                    videoUrl = this.isIE() ? streamInfo.WS_FLV || "" : streamInfo.FLV || "";
                }
                this.audio = !!streamInfo.AudioEnable;
                this.videoUrl = videoUrl;
            }

            this.sending = true;
            await this.$store.dispatch("connect", {
                topic: "下发视频参数属性配置",
            });
            var items = "";
            for(var item of this.VideoParamAttribute.Item) {
                items += `<Item>
<StreamNumber>${item.StreamNumber}</StreamNumber>
<VideoFormat>${item.VideoFormat}</VideoFormat>
<Resolution>${item.Resolution}</Resolution>
<FrameRate>${item.FrameRate}</FrameRate>
<BitRateType>${item.BitRateType}</BitRateType>
<VideoBitRate>${item.VideoBitRate}</VideoBitRate>
</Item>\r\n`;
            }
            const xml = `<VideoParamAttribute Num="${this.VideoParamAttribute.Item.length}">
${items}</VideoParamAttribute>`;
            $.post("/api/v1/control/deviceconfig", {
                serial: this.$store.state.serial,
                code: this.$store.state.code,
                VideoParamAttribute: xml,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "视频参数属性配置成功"
                })
                this.$store.commit("updateResult", ret);
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).fail(xhr => {
                this.$store.commit("updateResult", null);
            }).always(() => {
                this.sending = false;
            })
        },
        async sendOther() {
            const select = this.$el.querySelector('select[name="config_type"]');
            let label = this.config_type;
            if (select) {
                const option = Array.from(select.options).find(opt => opt.value === this.config_type);
                if (option) label = option.text;
            }
            if(!label.endsWith("配置")) {
                label += "配置";
            }
            this.sending = true;
            await this.$store.dispatch("connect", {
                topic: `下发${label}` ,
            });

            var otherParam = {
                serial: this.$store.state.serial,
                code: this.$store.state.code,
            }
            if(this.config_type && this.OtherConfig) {
                otherParam[this.config_type] = this.OtherConfig;
            }
            $.post("/api/v1/control/deviceconfig", otherParam).then(ret => {
                this.$message({
                    type: "success",
                    message: "其他参数配置成功"
                })
                this.$store.commit("updateResult", ret);
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).fail(xhr => {
                this.$store.commit("updateResult", null);
            }).always(() => {
                this.sending = false;
            })
        },
        formatStreamNumber(num) {
            if(!num || num < 0) {
                return "主码流";
            }
            if(num == 1) {
                return "子码流";
            }
            return `子码流${num}`;
        },
    }
}
</script>

<style lang="less" scoped>
.el-radio-group {
    margin-top: 9px;

    .el-radio + .el-radio {
        margin-left: 15px;
    }
}
</style>
