layui.define(function(exports){ //提示：模块也可以依赖其它模块，如：layui.define('layer', callback);
  var obj = {
    timestampToTime: function(timestamp) {
        var date = new Date(timestamp * 1000);//时间戳为10位需*1000，时间戳为13位的话不需乘1000
        Y = date.getFullYear() + '-';
        M = (date.getMonth()+1 < 10 ? '0'+(date.getMonth()+1) : date.getMonth()+1) + '-';
        D = date.getDate() + ' ';
        h = date.getHours() + ':';
        m = date.getMinutes() + ':';
        s = date.getSeconds();
        return Y+M+D+h+m+s;
    },
    UTCToTime: function(UTCTime) {
        var date = new Date(UTCTime);//时间戳为10位需*1000，时间戳为13位的话不需乘1000
        Y = date.getFullYear() + '-';
        M = (date.getMonth()+1 < 10 ? '0'+(date.getMonth()+1) : date.getMonth()+1) + '-';
        D = date.getDate() + ' ';
        h = date.getHours() + ':';
        m = date.getMinutes() + ':';
        s = date.getSeconds();
        return Y+M+D+h+m+s;
    },
    getRequest: function() {
        var url = decodeURIComponent(location.search); //获取url中"?"符后的字串
        var theRequest = {
            add:function(key, val) {
                 this[key] = val;
                 return this;
            },
            search:function()  {
                var query = "?";
                for (var i in this) {
                    if (typeof this[i] == "function") {
                        continue;
                    }
                    if (query == "?") {
                        query += i +"=" + this[i];
                    } else {
                        query += "&" + i +"=" + this[i];
                    }  
                }
               return query;
            }
        };
        if (url.indexOf("?") != -1) {
            var str = url.substr(1);
            strs = str.split("&");
            for (var i = 0; i < strs.length; i++) {
                theRequest[strs[i].split("=")[0]] = (strs[i].split("=")[1]);
            }
        }

        return theRequest;
    }
  };
  
  //输出common
  exports('common', obj);
});  