<template>
<div>
    <div class="input-area form-horizontal">
        <div class="form-group">
            <label class="col-sm-4 control-label">预置位操作
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <el-radio-group size="mini" v-model.trim="command" :disabled="loading">
                    <el-radio label="query">查询</el-radio>
                    <el-radio label="goto">调用</el-radio>
                    <el-radio label="set">设置</el-radio>
                    <el-radio label="remove">删除</el-radio>
                </el-radio-group>
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">预置位编号
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <el-input-number size="small" v-model.number="preset" :min="1" :max="255"></el-input-number>
                <span style="margin-left: 10px; color: #999;">{{presetName}}</span>
            </div>
        </div>
        <div class="form-group" v-show="command == 'set'">
            <label class="col-sm-4 control-label">预置位名称
            </label>
            <div class="col-sm-7">
                <input type="text" v-model.trim="name" class="form-control" style="width: 130px;">
            </div>
        </div>
        <div class="form-group">
            <div class="col-sm-offset-4 col-sm-7">
                <button role="button" class="btn btn-primary" @click.prevent="send" :disabled="loading">发送<span v-show="loading">...</span></button>
            </div>
        </div>
    </div>

    <div class="play-area" v-if="videoUrl" v-show="command != 'query'">
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
            command: "query",
            preset: 1,
            name: "",
            videoUrl: "",
            audio: false,
        }
    },
    computed: {
        ...mapState(["result"]),
        presetMap() {
            var list = [];
            if(this.result) {
                list = this.result.PresetItemList||[];
            }
            return list.reduce((pval, cval) => {
                pval[cval.PresetID] = cval;
                return pval;
            }, {});
        },
        presetName() {
            var preset = this.presetMap[this.preset];
            if(preset) {
                return preset.PresetName;
            }
            return "";
        },
    },
    watch: {
        command: function(newVal, oldVal) {
            if(newVal != "query") {
                this.$store.dispatch("play").then(streamInfo => {
                    if(streamInfo) {
                        var videoUrl = streamInfo.HLS || "";
                        if(this.flvSupported()) {
                            videoUrl = this.isIE() ? streamInfo.WS_FLV || "" : streamInfo.FLV || "";
                        }
                        this.audio = !!streamInfo.AudioEnable;
                        this.videoUrl = videoUrl;
                    }
                }).catch(() => {});
            }
            if(newVal == "set") {
                this.name = this.presetName;
            }
        },
        preset: function(newVal, oldVal) {
            var preset = this.presetMap[newVal];
            if(preset) {
                this.name = preset.PresetName;
            } else {
                this.name = "";
            }
        }
    },
    components: { LivePlayer },
    methods: {
        send() {
            if(this.command == "query") {
                this.query();
            } else {
                this.control();
            }
        },
        async query() {
            this.loading = true;
            await this.$store.dispatch("connect", {
                topic: "预置位查询",
            });

            $.get("/api/v1/device/fetchpreset", {
                serial: this.$store.state.serial,
                code: this.$store.state.code,
                fill: false,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "预置位查询成功"
                })
                this.$store.commit("updateResult", ret);
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).always(() => {
                this.loading = false;
            })
        },
        async control() {
            this.loading = true;
            await this.$store.dispatch("connect");

            $.get("/api/v1/control/preset", {
                serial: this.$store.state.serial,
                code: this.$store.state.code,
                command: this.command,
                preset: this.preset,
                name: this.name,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "预置位操作成功"
                })
                if(this.command == "set" || this.command == "remove") {
                    this.$store.commit("updateResult", null);
                }
            }).always(() => {
                this.loading = false;
            })
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

.play-area {
    margin: 50px auto;
    margin-bottom: 0;
}
</style>
