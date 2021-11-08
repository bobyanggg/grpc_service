# grpc_service

設計一個比價網 gRPC Service，提供以下 API，讓用戶同時搜尋不同購物網站：
service YourService {
  rpc Search (Request) returns (stream Result) {}
}
其中，Request 是商品的關鍵字；Result 包含名稱、價錢、圖片連結、商品連結
需求
* 至少兩個購物網站
* Exported functions / variables 要有註解
* 要寫 Unit test
* 運用 worker 技巧，並提供 flag 設定單一網站最多 workers 數量
* 搜尋結果有多頁時，也要爬下來
* 善用 interface 抽換底層實作，讓 code 具有延展性，並更容易測試
* 程式被中斷時，worker 必須把手上任務完成才結束
加分
* 寫一個對應的的前端 web application
* 若等每一頁的結果都收集好才一起回傳，User 可能會等很久，UX 不佳。請嘗試 streaming 回傳
* 運用 Database 建立 cache 機制，特定期限內 user 再次搜尋相同關鍵字，就不用再爬一次
* 但請勿 hard-code DB 連線資訊
* 結構化的 log 有助於了解程式運作情形，有效 debug
提示
* 善用 opensource libraries
* 避免頻繁的 request 導致購物網站誤判 DDoS 攻擊，請注意 rate limit
* 避免商品太多，可設定數量上限，或是過濾缺貨商品
合作很重要，請把專案上傳 Github（GitLab 或其他 repostory provider 亦可）