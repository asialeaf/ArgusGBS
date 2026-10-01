<template>
  <input type="text" placeholder="选择日期">
</template>

<script>
import moment from 'moment'

export default {
    data() {
        return {
            flags : {},
            defaultFlags : "0000000000000000000000000000000"
        }
    },
    props: {
        serial: {
            default: ""
        },
        code: {
            default: ""
        },
        token: {
            default: ""
        },
        day: {
            default: () => moment().format("YYYYMMDD")
        }
    },
    mounted() {
        $(this.$el).datepicker({
            language : "zh-CN",
            autoclose : true,
            format : "yyyy-mm-dd",
            todayHighlight : false,
            beforeShowDay: date => {
                var period = moment(date).format("YYYYMM");
                var flags = this.getFlagsByMonth(period);
                var d = moment(date).date();
                var flag = flags[d - 1];
                if(flag == '0'){
                    return {classes : "text-gray", tooltip : "没有录像", enabled : false};
                } else {
                    return {classes : "text-primary text-bold", tooltip : "有录像"};
                }
            }
        }).datepicker("setDate", moment(this.day,"YYYYMMDD").toDate()).on("changeDate",e => {
            this.$emit("update:day", moment(e.date).format('YYYYMMDD'));
        })
        this.update();
    },
    watch: {
        day: function(val){
            let d = $(this.$el).datepicker("getDate");
            if(val != moment(d).format("YYYYMMDD")){
                $(this.$el).datepicker("setDate", moment(val,"YYYYMMDD").toDate());
            }
        },
        id: function(val) {
            this.update();
        }
    },
    computed: {
        id() {
            return `${this.serial}_${this.code}`;
        }
    },
    methods: {
        update() {
            this.getFlagMap();
        },
        getFlagMap() {
            if(!this.serial || !this.code) return;
            $.get('/api/v1/cloudrecord/queryflags',{
                serial: this.serial,
                code: this.code,
                token: this.token,
            }).then(data => {
                this.flags = data;
                $(this.$el).datepicker('update');
            })
        },
        getFlagsByMonth(period) {
            return this.flags[period] || this.defaultFlags;
        }
    }
}
</script>

<style lang="less">
.datepicker {
    z-index: 9999 !important;
}
</style>
