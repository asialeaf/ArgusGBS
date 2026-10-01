<template>
<div>
    <el-tabs v-model="tab_label">
        <el-tab-pane label="目录订阅" name="catalog">
            <div class="catalog-area form-horizontal">
                <div class="form-group">
                    <label class="control-label col-sm-4">订阅时长(秒)
                        <span class="text-danger">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="number" class="form-control" v-model.number="catalogExpires" min="1"/>
                    </div>
                </div>
                <div class="form-group">
                    <div class="col-sm-offset-4 col-sm-7">
                        <button role="button" class="btn btn-primary" @click.prevent="subCatalog(true)" :disabled="sending">订阅<span v-show="sending">...</span></button>
                        <button role="button" class="btn btn-info" @click.prevent="subCatalog(false)" :disabled="sending">取消订阅<span v-show="sending">...</span></button>
                    </div>
                </div>
            </div>
        </el-tab-pane>
        <el-tab-pane label="移动位置订阅" name="mobile">
            <div class="mobile-area form-horizontal">
                <div class="form-group">
                    <label class="control-label col-sm-4">订阅时长(秒)
                        <span class="text-danger">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="number" class="form-control" v-model.number="mobileExpires" min="1"/>
                    </div>
                </div>
                <div class="form-group">
                    <label class="control-label col-sm-4">上报间隔(秒)
                        <span class="text-danger">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="number" class="form-control" v-model.number="interval" min="1"/>
                    </div>
                </div>
                <div class="form-group">
                    <div class="col-sm-offset-4 col-sm-7">
                        <button role="button" class="btn btn-primary" @click.prevent="subMobile(true)" :disabled="sending">订阅<span v-show="sending">...</span></button>
                        <button role="button" class="btn btn-info" @click.prevent="subMobile(false)" :disabled="sending">取消订阅<span v-show="sending">...</span></button>
                    </div>
                </div>
            </div>
        </el-tab-pane>
        <el-tab-pane label="PTZ精准状态订阅" name="ptz">
            <div class="ptz-area form-horizontal">
                <div class="form-group">
                    <label class="control-label col-sm-4">订阅时长(秒)
                        <span class="text-danger">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="number" class="form-control" v-model.number="ptzExpires" min="1"/>
                    </div>
                </div>
                <div class="form-group">
                    <div class="col-sm-offset-4 col-sm-7">
                        <button role="button" class="btn btn-primary" @click.prevent="subPtz(true)" :disabled="sending">订阅<span v-show="sending">...</span></button>
                        <button role="button" class="btn btn-info" @click.prevent="subPtz(false)" :disabled="sending">取消订阅<span v-show="sending">...</span></button>
                    </div>
                </div>
            </div>
        </el-tab-pane>
        <el-tab-pane label="报警订阅" name="alarm">
            <div class="alarm-area form-horizontal">
                <div class="form-group">
                    <label class="control-label col-sm-4">订阅时长(秒)
                        <span class="text-danger">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="number" class="form-control" v-model.number="alarmExpires" min="1"/>
                    </div>
                </div>
                <div class="form-group">
                    <label class="control-label col-sm-4">报警方式
                    </label>
                    <div class="col-sm-7">
                        <select class="form-control" v-model.trim="alarmMethod">
                            <option value="">不选</option>
                            <option value="1">1-电话报警</option>
                            <option value="2">2-设备报警</option>
                            <option value="3">3-短信报警</option>
                            <option value="4">4-GPS报警</option>
                            <option value="5">5-视频报警</option>
                            <option value="6">6-设备故障报警</option>
                            <option value="7">7-其他报警</option>
                        </select>
                    </div>
                </div>
                <div class="form-group">
                    <label class="control-label col-sm-4">报警类型
                    </label>
                    <div class="col-sm-7">
                        <select class="form-control" v-model.trim="alarmType">
                            <option v-for="type in alarmTypes" :key="type.value" :value="type.value">{{ type.label }}</option>
                        </select>
                    </div>
                </div>
                <div class="form-group">
                    <div class="col-sm-offset-4 col-sm-7">
                        <button role="button" class="btn btn-primary" @click.prevent="subAlarm(true)" :disabled="sending">订阅<span v-show="sending">...</span></button>
                        <button role="button" class="btn btn-info" @click.prevent="subAlarm(false)" :disabled="sending">取消订阅<span v-show="sending">...</span></button>
                    </div>
                </div>
            </div>
        </el-tab-pane>
    </el-tabs>
</div>
</template>
<script>
import { mapState } from "vuex";
const regNotify = /<Notify>[\s\S]*<CmdType>(.+)<\/CmdType>[\s\S]*<\/Notify>/m;
export default {
    data() {
        return {
            tab_label: "catalog",
            sending: false,
            catalogExpires: 60,
            mobileExpires: 60,
            interval: 5,
            ptzExpires: 60,
            alarmExpires: 60,
            alarmMethod: "",
            alarmType: "",
            queue: [],
            result: [],
            timer: 0,
        };
    },
    computed: {
        ...mapState(["msg"]),
        alarmTypes() {
            var list = [{ value: "", label: "不选" }];
            switch(this.alarmMethod) {
                case "2": // 设备报警
                    list.push({ value: "1", label: "1-视频丢失报警" },{ value: "2", label: "2-设备防拆报警" },
                        { value: "3", label: "3-存储设备磁盘满报警" }, { value: "4", label: "4-设备高温报警" },
                        { value: "5", label: "5-设备低温报警" });
                    break;
                case "5": // 视频报警
                    list.push({ value: "1", label: "1-人工视频报警" },{ value: "2", label: "2-运动目标检测报警" },
                        { value: "3", label: "3-遗留物检测报警" }, { value: "4", label: "4-物体移除检测报警" },
                        { value: "5", label: "5-绊线检测报警" }, { value: "6", label: "6-入侵检测报警" },
                        { value: "7", label: "7-逆行检测报警" }, { value: "8", label: "8-徘徊检测报警" },
                        { value: "9", label: "9-流量统计报警" }, { value: "10", label: "10-密度检测报警" },
                        { value: "11", label: "11-视频异常检测报警" }, { value: "12", label: "12-快速移动报警" },
                        { value: "13", label: "13-图像遮挡报警" });
                    break;
                case "6": // 设备故障报警
                    list.push({ value: "1", label: "1-存储设备磁盘故障报警" },{ value: "2", label: "2-存储设备风扇故障报警" });
                    break;
            }
            return list;
        },
    },
    watch: {
        alarmMethod: function(newVal, oldVal) {
            this.alarmType = "";
        },
        msg: function(newVal, oldVal) {
            if(!newVal || newVal.length <= oldVal.length) return;
            var last = newVal.substr(oldVal.length);
            if(last.startsWith("<<<<<<")) {
                var match = regNotify.exec(last);
                if(match && match.length > 1 && match[1]) {
                    switch(match[1]) {
                        case "Catalog":
                            this.pushResult(`收到目录通知`);
                            break;
                        case "MobilePosition":
                            this.pushResult(`收到移动位置通知`);
                            break;
                        case "PTZPosition":
                            this.pushResult(`收到PTZ精确状态通知`);
                            break;
                        case "Alarm":
                            this.pushResult(`收到报警通知`);
                            break;
                    }
                    return
                }
            }
        }
    },
    methods: {
        async subCatalog(sub) {
            this.sending = true;
            await this.$store.dispatch("connect");

            $.post("/api/v1/control/devicesubscribe", {
                serial: this.$store.state.serial,
                code: this.$store.state.code||this.$store.state.serial,
                type: "catalog",
                expires: sub ? this.ptzExpires : 0,
            }).then(ret => {
                var msg = (sub ? "订阅" : "取消订阅") + "目录";
                this.$message({
                    type: "success",
                    message: `${msg}成功`,
                });
                this.pushResult(msg);
                // if(!sub) {
                //     setTimeout(() => {
                //         this.$store.dispatch("disconnect");
                //     }, 1000);
                // }
            }).always(() => {
                this.sending = false;
            });
        },
        async subMobile(sub) {
            this.sending = true;
            await this.$store.dispatch("connect");

            $.post("/api/v1/control/devicesubscribe", {
                serial: this.$store.state.serial,
                code: this.$store.state.code||this.$store.state.serial,
                type: "mobile",
                expires: sub ? this.mobileExpires : 0,
                interval: this.interval,
            }).then(ret => {
                var msg = (sub ? "订阅" : "取消订阅") + "移动位置";
                this.$message({
                    type: "success",
                    message: `${msg}成功`,
                });
                this.pushResult(msg);
                // if(!sub) {
                //     setTimeout(() => {
                //         this.$store.dispatch("disconnect");
                //     }, 1000);
                // }
            }).always(() => {
                this.sending = false;
            });
        },
        async subPtz(sub) {
            this.sending = true;
            await this.$store.dispatch("connect");

            $.post("/api/v1/control/devicesubscribe", {
                serial: this.$store.state.serial,
                code: this.$store.state.code||this.$store.state.serial,
                type: "ptz",
                expires: sub ? this.ptzExpires : 0,
            }).then(ret => {
                var msg = (sub ? "订阅" : "取消订阅") + "PTZ精准状态";
                this.$message({
                    type: "success",
                    message: `${msg}成功`,
                });
                this.pushResult(msg);
                // if(!sub) {
                //     setTimeout(() => {
                //         this.$store.dispatch("disconnect");
                //     }, 1000);
                // }
            }).always(() => {
                this.sending = false;
            });
        },
        async subAlarm(sub) {
            this.sending = true;
            await this.$store.dispatch("connect");

            $.post("/api/v1/control/devicesubscribe", {
                serial: this.$store.state.serial,
                code: this.$store.state.code||this.$store.state.serial,
                type: "alarm",
                alarm_method: this.alarmMethod,
                alarm_type: this.alarmType,
                expires: sub ? this.alarmExpires : 0,
            }).then(ret => {
                var msg = (sub ? "订阅" : "取消订阅") + "报警";
                this.$message({
                    type: "success",
                    message: `${msg}成功`,
                });
                this.pushResult(msg);
                // if(!sub) {
                //     setTimeout(() => {
                //         this.$store.dispatch("disconnect");
                //     }, 1000);
                // }
            }).always(() => {
                this.sending = false;
            });
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
                if(!this.$store.state.result) {
                    if(this.$store.state.ws) {
                        this.$store.commit("updateResult", this.result);
                    } else {
                        this.result.length = 0;
                    }
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