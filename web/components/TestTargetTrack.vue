<template>
<div>
    <div class="input-area form-horizontal">
        <div class="form-group">
            <label class="col-sm-4 control-label">全景通道
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <el-select filterable clearable remote :remote-method="loadChannel" :loading="channelLoading" v-model="device_id2"
                    @focus="!device_id2 && loadChannel();" ref="selectChannel"
                    name="device_id2" placeholder="可搜索" size="small" style="width:100%;">
                    <el-option v-for="item in channels" :key="item.ID" :label="item.ID" :value="item.ID">
                    </el-option>
                </el-select>
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">目标跟踪命令
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <el-radio-group size="mini" v-model.trim="target_track" :disabled="sending">
                    <el-radio label="Auto">自动</el-radio>
                    <el-radio label="Manual">手动</el-radio>
                    <el-radio label="Stop">停止</el-radio>
                </el-radio-group>
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">全景窗口长
            </label>
            <div class="col-sm-7">
                <input type="number" min="0" v-model="length" class="form-control" style="width: 100px;">
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">全景窗口宽
            </label>
            <div class="col-sm-7">
                <input type="number" min="0" v-model="width" class="form-control" style="width: 100px;">
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">跟踪框中心X
            </label>
            <div class="col-sm-7">
                <input type="number" min="0" v-model="midpointx" class="form-control" style="width: 100px;">
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">跟踪框中心Y
            </label>
            <div class="col-sm-7">
                <input type="number" min="0" v-model="midpointy" class="form-control" style="width: 100px;">
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">跟踪框长
            </label>
            <div class="col-sm-7">
                <input type="number" min="0" v-model="lengthx" class="form-control" style="width: 100px;">
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">跟踪框宽
            </label>
            <div class="col-sm-7">
                <input type="number" min="0" v-model="lengthy" class="form-control" style="width: 100px;">
            </div>
        </div>
        <div class="form-group">
            <div class="col-sm-offset-4 col-sm-7">
                <button role="button" class="btn btn-primary" @click.prevent="send" :disabled="sending">发送<span v-show="sending">...</span></button>
            </div>
        </div>
    </div>

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
            channelLoading: false,
            channels: [],
            sending: false,
            target_track: "Auto",
            device_id2: "",
            length: "",
            width: "",
            midpointx: "",
            midpointy: "",
            lengthx: "",
            lengthy: "",
            videoUrl: "",
            audio: false,
        }
    },
    computed: {
        ...mapState(["serial"]),
    },
    watch: {
        serial: function(newVal, oldVal) {
            this.device_id2 = "";
            this.loadChannel();
        }
    },
    components: { LivePlayer },
    methods: {
        async send() {
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
            await this.$store.dispatch("connect");

            $.post("/api/v1/control/targettrack", {
                serial: this.$store.state.serial,
                code: this.$store.state.code,
                target_track: this.target_track,
                device_id2: this.device_id2,
                length: this.length,
                width: this.width,
                midpointx: this.midpointx,
                midpointy: this.midpointy,
                lengthx: this.lengthx,
                lengthy: this.lengthy,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "目标跟踪命令控制成功"
                })
                if(this.target_track == "Stop") {
                    setTimeout(() => {
                        this.$store.dispatch("disconnect");
                    }, 1000);
                }
            }).always(() => {
                this.sending = false;
            })
        },
        loadChannel(q = "") {
            if(!this.serial) {
                this.channels.length = 0;
                return;
            }
            this.channelLoading = true;
            $.get("/api/v1/device/channellist", {
                q: q == this.device_id2 ? '' : q,
                serial: this.serial,
                limit: 10,
            }).then(ret => {
                this.channels = ret.ChannelList || [];
                if(this.serial && !this.device_id2 && !q && this.channels.length > 0) {
                    this.device_id2 = this.channels[0].ID;
                }
            }).always(() => {
                this.channelLoading = false;
            })
        },
    },
    mounted() {
        // this.loadChannel();
    },
}
</script>

<style lang="less" scoped>
.el-radio-group {
    margin-top: 9px;

    .el-radio + .el-radio {
        margin-left: 15px;
    }
}

.play-area {
    margin: 20px auto;
}
</style>
