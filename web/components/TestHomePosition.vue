<template>
<div>
    <div class="input-area form-horizontal">
        <div class="form-group">
            <label class="col-sm-4 control-label">看守位操作
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <el-radio-group size="mini" v-model.trim="enabled" :disabled="sending">
                    <el-radio :label="true">开启</el-radio>
                    <el-radio :label="false">关闭</el-radio>
                </el-radio-group>
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">归位时间(秒)
            </label>
            <div class="col-sm-7">
                <el-input-number size="small" v-model="resettime" :min="1"></el-input-number>
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">预置位编号
            </label>
            <div class="col-sm-7">
                <el-input-number size="small" v-model="preset" :min="1" :max="255"></el-input-number>
            </div>
        </div>
        <div class="form-group">
            <div class="col-sm-offset-4 col-sm-7">
                <button role="button" class="btn btn-primary" @click.prevent="send" :disabled="sending">发送<span v-show="sending">...</span></button>
                <button role="button" class="btn btn-info" @click.prevent="query" :disabled="loading">查询<span v-show="loading">...</span></button>
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
import LivePlayer from "@liveqing/liveplayer";

export default {
    data() {
        return {
            sending: false,
            loading: false,
            enabled: false,
            resettime: 5,
            preset: 1,
            videoUrl: "",
            audio: false,
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

            $.post("/api/v1/control/homeposition", {
                serial: this.$store.state.serial,
                code: this.$store.state.code,
                enabled: this.enabled,
                resettime: this.resettime,
                presetindex: this.preset,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "看守位操作成功"
                })
                this.$store.commit("updateResult", ret);
            }).fail(ret => {
                this.$store.commit("updateResult", null);
            }).always(() => {
                this.sending = false;
            })
        },
        async query() {
            this.loading = true;
            await this.$store.dispatch("connect");

            $.get("/api/v1/device/fetchhomeposition", {
                serial: this.$store.state.serial,
                code: this.$store.state.code,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "看守位查询成功"
                })
                this.$store.commit("updateResult", ret);
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).always(() => {
                this.loading = false;
            })
        },
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
    margin-bottom: 0;
}
</style>
