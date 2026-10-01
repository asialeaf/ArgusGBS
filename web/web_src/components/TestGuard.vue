<template>
<div>
    <div class="input-area form-horizontal">
        <div class="form-group">
            <label class="col-sm-4 control-label">布撤防
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <el-radio-group size="mini" v-model.trim="command" :disabled="sending">
                    <el-radio label="setguard">布防</el-radio>
                    <el-radio label="resetguard">撤防</el-radio>
                </el-radio-group>
            </div>
        </div>
        <div class="form-group">
            <div class="col-sm-offset-4 col-sm-7">
                <button role="button" class="btn btn-primary" @click.prevent="send" :disabled="sending">发送<span v-show="sending">...</span></button>
            </div>
        </div>
    </div>
</div>
</template>

<script>
import { mapState } from "vuex";
const regNotify = /<Notify>[\s\S]*<CmdType>(.+)<\/CmdType>[\s\S]*<\/Notify>/m;
export default {
    data() {
        return {
            sending: false,
            command: "setguard",
            queue: [],
            result: [],
            timer: 0,
        }
    },
    computed: {
        ...mapState(["msg"]),
    },
    watch: {
        msg: function(newVal, oldVal) {
            if(!newVal || newVal.length <= oldVal.length) return;
            var last = newVal.substr(oldVal.length);
            if(last.startsWith("<<<<<<")) {
                var match = regNotify.exec(last);
                if(match && match.length > 1 && match[1]) {
                    switch(match[1]) {
                        case "Alarm":
                            this.pushResult(`收到报警消息`);
                            break;
                    }
                    return
                }
            }
        }
    },
    methods: {
        async send() {
            this.sending = true;
            await this.$store.dispatch("connect");

            $.post("/api/v1/control/guard", {
                serial: this.$store.state.serial,
                code: this.$store.state.code||this.$store.state.serial,
                command: this.command,
            }).then(ret => {
                switch(this.command) {
                    case "setguard":
                        this.$message({
                            type: "success",
                            message: "布防操作成功"
                        })
                        break;
                    case "resetguard":
                        this.$message({
                            type: "success",
                            message: "撤防操作成功"
                        })
                        break;
                    default:
                        this.$message({
                            type: "success",
                            message: "布撤防操作成功"
                        })
                        break;
                }
                this.$store.commit("updateResult", ret);
                if(this.command == "resetguard") {
                    setTimeout(() => {
                        this.$store.dispatch("disconnect");
                    }, 1000);
                }
            }).fail(ret => {
                this.$store.commit("updateResult", null);
            }).always(() => {
                this.sending = false;
            })
        },
        pushResult(result) {
            if(!result) return;
            this.queue.push(result);
            if(this.timer) {
                clearTimeout(this.timer);
                this.timer = 0;
            }
            this.timer = setTimeout(() => {
                this.result.push(...this.queue);
                this.queue.length = 0;
                if(this.$store.state.ws) {
                    this.$store.commit("updateResult", this.result);
                } else {
                    this.result.length = 0;
                }
            }, 1000);
        },
        reset() {
            if(this.timer) {
                clearTimeout(this.timer);
                this.timer = 0;
            }
        },
        beforeUnload(event) {
            this.reset();
            event.returnValue = '';
        },
    },
    mounted() {
        $(window).on("beforeunload", this.beforeUnload);
    },
    beforeDestroy() {
        $(window).off("beforeunload", this.beforeUnload);
        this.reset();
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
</style>
