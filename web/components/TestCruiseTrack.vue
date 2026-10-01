<template>
<div>
    <el-tabs v-model="tab_label">
        <el-tab-pane label="巡航轨迹列表查询" name="queryList">
            <div class="query-list-area form-horizontal">
                <div class="form-group">
                    <div class="col-sm-offset-4 col-sm-7">
                        <button role="button" class="btn btn-primary" @click.prevent="queryList" :disabled="listLoading">发送<span v-show="listLoading">...</span></button>
                    </div>
                </div>
            </div>
        </el-tab-pane>
        <el-tab-pane label="巡航轨迹查询" name="queryTrack">
            <div class="query-track-area form-horizontal">
                <div class="form-group">
                    <label class="control-label col-sm-4">巡航轨迹编号
                        <span class="text-danger">*</span>
                    </label>
                    <div class="col-sm-7">
                        <input type="number" class="form-control" v-model.number="trackNumber" min="0"/>
                    </div>
                </div>
                <div class="form-group">
                    <div class="col-sm-offset-4 col-sm-7">
                        <button role="button" class="btn btn-primary" @click.prevent="queryTrack" :disabled="trackLoading">发送<span v-show="trackLoading">...</span></button>
                    </div>
                </div>
            </div>
        </el-tab-pane>
    </el-tabs>
</div>
</template>
<script>
export default {
    data() {
        return {
            tab_label: "queryList",
            listLoading: false,
            trackLoading: false,
            trackNumber: 0,
        };
    },
    methods: {
        async queryList() {
            this.listLoading = true;
            await this.$store.dispatch("connect", {
                topic: "巡航轨迹列表查询",
            });

            $.get("/api/v1/device/fetchcruisetracklist", {
                serial: this.$store.state.serial,
                code: this.$store.state.code,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "巡航轨迹列表查询成功"
                });
                this.$store.commit("updateResult", ret);
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).always(() => {
                this.listLoading = false;
            });
        },
        async queryTrack() {
            this.trackLoading = true;
            await this.$store.dispatch("connect", {
                topic: "巡航轨迹查询",
            });

            $.get("/api/v1/device/fetchcruisetrack", {
                serial: this.$store.state.serial,
                code: this.$store.state.code,
                number: this.trackNumber,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "巡航轨迹查询成功"
                });
                this.$store.commit("updateResult", ret);
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).always(() => {
                this.trackLoading = false;
            });
        },
    },
};
</script>