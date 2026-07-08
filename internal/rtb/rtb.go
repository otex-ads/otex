package rtb

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"adnet/internal/store/redis"
)

// RTB (Real-Time Bidding) implementation for programmatic ad buying
// Supports OpenRTB 2.5 standard

type RTBEngine struct {
	redis            *redis.Client
	bidders          map[string]Bidder
	bidTimeout       time.Duration
	maxConcurrentBids int
	mu               sync.RWMutex
}

type Bidder struct {
	ID          string
	Name        string
	Endpoint    string
	AuthToken   string
	Enabled     bool
	AvgResponse time.Duration
	WinRate     float64
}

type BidRequest struct {
	ID        string      `json:"id"`
	Imp       []Impression `json:"imp"`
	Site      Site        `json:"site"`
	App       *App        `json:"app,omitempty"`
	Device    Device      `json:"device"`
	User      *User       `json:"user,omitempty"`
	Test      int         `json:"test"`
	At        int         `json:"at"` // Auction type: 1=first price, 2=second price
	TMax      int         `json:"t_max"` // Max time for bids in ms
	AllImps   int         `json:"allimps"` // 0 = respond with bid for any imp, 1 = respond with bid for all imps
	Cur       []string    `json:"cur"` // Currencies
	WLang     []string    `json:"wlang"` // Languages
	BSeat     []string    `json:"bseat"` // Buyer seats
}

type Impression struct {
	ID       string       `json:"id"`
	TagID    string       `json:"tagid"`
	BidFloor float64      `json:"bidfloor"`
	BidFloorCur string    `json:"bidfloorcur"`
	Format   []Format     `json:"format,omitempty"`
	Banner   *Banner      `json:"banner,omitempty"`
	Video    *Video       `json:"video,omitempty"`
	Native   *Native      `json:"native,omitempty"`
	Secure   int          `json:"secure"`
	Ext      interface{}  `json:"ext,omitempty"`
}

type Format struct {
	W int `json:"w"`
	H int `json:"h"`
	MinW int `json:"minw,omitempty"`
	MinH int `json:"minh,omitempty"`
	Ratio int `json:"ratio,omitempty"`
}

type Banner struct {
	W        int      `json:"w"`
	H        int      `json:"h"`
	WMax     int      `json:"wmax,omitempty"`
	HMax     int      `json:"hmax,omitempty"`
	WMin     int      `json:"wmin,omitempty"`
	HMin     int      `json:"hmin,omitempty"`
	BType    []int    `json:"btype,omitempty"`
	BAttr    []int    `json:"battr,omitempty"`
	Pos      int      `json:"pos,omitempty"`
	API      []int    `json:"api,omitempty"`
	ID       string   `json:"id,omitempty"`
	VCM      int      `json:"vcm,omitempty"`
	Ext      interface{} `json:"ext,omitempty"`
}

type Video struct {
	MIMEs           []string `json:"mimes"`
	MinDuration     int      `json:"minduration"`
	MaxDuration     int      `json:"maxduration"`
	Protocols       []int    `json:"protocols"`
	W               int      `json:"w"`
	H               int      `json:"h"`
	StartDelay      int      `json:"startdelay"`
	Placement       int      `json:"placement"`
	Linearity       int      `json:"linearity"`
	MinBitrate      int      `json:"minbitrate"`
	MaxBitrate      int      `json:"maxbitrate"`
	Delivery        []int    `json:"delivery"`
	PlaybackMethod  []int    `json:"playbackmethod"`
	PlaybackEnd     int      `json:"playbackend"`
	DeliveryType    []int    `json:"deliverytype"`
	API             []int    `json:"api"`
	CompanionAd     []Banner `json:"companionad"`
	Ext             interface{} `json:"ext,omitempty"`
}

type Native struct {
	Request string `json:"request"`
	Ver     string `json:"ver"`
	BAttr   []int  `json:"battr"`
	API     []int  `json:"api"`
	Ext     interface{} `json:"ext,omitempty"`
}

type Site struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Domain      string            `json:"domain"`
	Cat         []string          `json:"cat,omitempty"`
	SectionCat  []string          `json:"sectioncat,omitempty"`
	PageCat     []string          `json:"pagecat,omitempty"`
	Page        string            `json:"page,omitempty"`
	Ref         string            `json:"ref,omitempty"`
	Search      string            `json:"search,omitempty"`
	Mobile      int               `json:"mobile"`
	PrivacyPolicy int             `json:"privacypolicy"`
	Publisher   *Publisher        `json:"publisher,omitempty"`
	Content     *Content          `json:"content,omitempty"`
	Keywords    string            `json:"keywords,omitempty"`
	Ext         interface{}       `json:"ext,omitempty"`
}

type App struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Bundle      string            `json:"bundle"`
	Domain      string            `json:"domain"`
	StoreURL    string            `json:"storeurl"`
	Cat         []string          `json:"cat,omitempty"`
	SectionCat  []string          `json:"sectioncat,omitempty"`
	PageCat     []string          `json:"pagecat,omitempty"`
	Ver         string            `json:"ver"`
	PrivacyPolicy int             `json:"privacypolicy"`
	Paid        int               `json:"paid"`
	Publisher   *Publisher        `json:"publisher,omitempty"`
	Content     *Content          `json:"content,omitempty"`
	Keywords    string            `json:"keywords,omitempty"`
	Ext         interface{}       `json:"ext,omitempty"`
}

type Publisher struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Domain   string   `json:"domain"`
	Cat      []string `json:"cat,omitempty"`
	Ext      interface{} `json:"ext,omitempty"`
}

type Content struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Series      string   `json:"series"`
	Season      string   `json:"season"`
	Episode     int      `json:"episode"`
	Artist      string   `json:"artist"`
	Genre       string   `json:"genre"`
	Album       string   `json:"album"`
	ISRC        string   `json:"isrc"`
	Producer    *Producer `json:"producer,omitempty"`
	URL         string   `json:"url"`
	Cat         []string `json:"cat,omitempty"`
	ProdQ       int      `json:"prodq"`
	Context     int      `json:"context"`
	LiveStream  int      `json:"livestream"`
	SourceRelationship int `json:"sourcerelationship"`
	Keywords    string   `json:"keywords"`
	Len         int      `json:"len"`
	Language    string   `json:"language"`
	QAGMediaRatings int `json:"qagmediaratings"`
	Ext         interface{} `json:"ext,omitempty"`
}

type Producer struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Domain   string   `json:"domain"`
	Cat      []string `json:"cat,omitempty"`
	Ext      interface{} `json:"ext,omitempty"`
}

type Device struct {
	UA           string   `json:"ua"`
	Geo          *Geo     `json:"geo,omitempty"`
	DNT          int      `json:"dnt"`
	LMT          int      `json:"lmt"`
	IP           string   `json:"ip"`
	IPv6         string   `json:"ipv6,omitempty"`
	DeviceType   int      `json:"devicetype"`
	Make         string   `json:"make"`
	Model        string   `json:"model"`
	OS           string   `json:"os"`
	OSV          string   `json:"osv"`
	HWV          string   `json:"hwv"`
	H            int      `json:"h"`
	W            int      `json:"w"`
	PPI          int      `json:"ppi"`
	PxRatio      float64  `json:"pxratio"`
	JS           int      `json:"js"`
	FlashVer     string   `json:"flashver"`
	Language     string   `json:"language"`
	Carrier      *Carrier `json:"carrier,omitempty"`
	MNC          string   `json:"mnc"`
	MCC          string   `json:"mcc"`
	ConnectionType int    `json:"connectiontype"`
	IFA          string   `json:"ifa"`
	LID          string   `json:"lid"`
	DIDMD5       string   `json:"didmd5"`
	DIDSHA1      string   `json:"didsha1"`
	DPIDMD5      string   `json:"dpidmd5"`
	DPIDSHA1     string   `json:"dpidsha1"`
	MACMD5       string   `json:"macmd5"`
	MACSHA1      string   `json:"macsha1"`
	Ext          interface{} `json:"ext,omitempty"`
}

type Geo struct {
	Lat          float64  `json:"lat"`
	Lon          float64  `json:"lon"`
	Type         int      `json:"type"`
	Accuracy     int      `json:"accuracy"`
	LastFix      int      `json:"lastfix"`
	IPService    int      `json:"ipservice"`
	Country      string   `json:"country"`
	Region       string   `json:"region"`
	RegionFIPS104 string  `json:"regionfips104"`
	Metro        string   `json:"metro"`
	City         string   `json:"city"`
	Zip          string   `json:"zip"`
	UTCOffset    int      `json:"utcoffset"`
	Ext          interface{} `json:"ext,omitempty"`
}

type Carrier struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Country      string   `json:"country"`
	MCCMNC       string   `json:"mccmnc"`
	Ext          interface{} `json:"ext,omitempty"`
}

type User struct {
	ID         string   `json:"id"`
	BuyerUID   string   `json:"buyeruid"`
	YOB        int      `json:"yob"`
	Gender     string   `json:"gender"`
	Keywords   string   `json:"keywords"`
	CustomData string   `json:"customdata"`
	Geo        *Geo     `json:"geo,omitempty"`
	Data       []Data   `json:"data,omitempty"`
	Ext        interface{} `json:"ext,omitempty"`
}

type Data struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Segment  []string `json:"segment"`
	Ext      interface{} `json:"ext,omitempty"`
}

type BidResponse struct {
	ID         string          `json:"id"`
	SeatBid    []SeatBid      `json:"seatbid"`
	BidID      string          `json:"bidid"`
	Cur        string          `json:"cur"`
	CustomData string          `json:"customdata,omitempty"`
	NBR        int             `json:"nbr"` // No Bid Reason
	Ext        interface{}     `json:"ext,omitempty"`
}

type SeatBid struct {
	Bid    []Bid          `json:"bid"`
	Seat   string         `json:"seat,omitempty"`
	Group  int            `json:"group,omitempty"`
	Ext    interface{}    `json:"ext,omitempty"`
}

type Bid struct {
	ID       string      `json:"id"`
	ImpID    string      `json:"impid"`
	Price    float64     `json:"price"`
	AdID     string      `json:"adid"`
	NURL     string      `json:"nurl,omitempty"`
	BURL     string      `json:"burl,omitempty"`
	LURL     string      `json:"lurl,omitempty"`
	AdM      int         `json:"adm,omitempty"`
	AdDomain string      `json:"addomain,omitempty"`
	Bundle   string      `json:"bundle,omitempty"`
	IURL     string      `json:"iurl,omitempty"`
	CID      string      `json:"cid,omitempty"`
	Crid     string      `json:"crid,omitempty"`
	Cat      []string    `json:"cat,omitempty"`
	Attr     []int       `json:"attr,omitempty"`
	API      int         `json:"api,omitempty"`
	Protocol int         `json:"protocol,omitempty"`
	QAGMediaRatings int   `json:"qagmediaratings,omitempty"`
	Language []string    `json:"language,omitempty"`
	DealID   string      `json:"dealid,omitempty"`
	W        int         `json:"w,omitempty"`
	H        int         `json:"h,omitempty"`
	WRatio   int         `json:"wratio,omitempty"`
	HRatio   int         `json:"hratio,omitempty"`
	Dur      int         `json:"dur,omitempty"`
	MIMEs    []string    `json:"mimes,omitempty"`
	Ext      interface{} `json:"ext,omitempty"`
}

func NewRTBEngine(redisClient *redis.Client) *RTBEngine {
	return &RTBEngine{
		redis:            redisClient,
		bidders:          make(map[string]Bidder),
		bidTimeout:       100 * time.Millisecond,
		maxConcurrentBids: 10,
	}
}

// RegisterBidder adds a new bidder to the RTB engine
func (e *RTBEngine) RegisterBidder(bidder Bidder) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.bidders[bidder.ID] = bidder
}

// RemoveBidder removes a bidder from the RTB engine
func (e *RTBEngine) RemoveBidder(bidderID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.bidders, bidderID)
}

// ConductAuction runs a real-time auction for an ad impression
func (e *RTBEngine) ConductAuction(ctx context.Context, bidRequest BidRequest) (*BidResponse, error) {
	e.mu.RLock()
	bidders := make([]Bidder, 0, len(e.bidders))
	for _, bidder := range e.bidders {
		if bidder.Enabled {
			bidders = append(bidders, bidder)
		}
	}
	e.mu.RUnlock()

	if len(bidders) == 0 {
		return &BidResponse{
			ID:  bidRequest.ID,
			NBR: 1, // No bidders available
		}, nil
	}

	// Send bid requests to all bidders concurrently
	bidResponses := make(chan *BidResponse, len(bidders))
	errors := make(chan error, len(bidders))

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, e.maxConcurrentBids)

	for _, bidder := range bidders {
		wg.Add(1)
		go func(b Bidder) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			resp, err := e.sendBidRequest(ctx, b, bidRequest)
			if err != nil {
				errors <- err
				return
			}
			bidResponses <- resp
		}(bidder)
	}

	// Wait for all bids or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(e.bidTimeout):
		log.Printf("RTB auction timeout for request %s", bidRequest.ID)
	}

	close(bidResponses)
	close(errors)

	// Collect bids
	var allBids []Bid
	for resp := range bidResponses {
		if resp != nil && len(resp.SeatBid) > 0 {
			for _, seatBid := range resp.SeatBid {
				allBids = append(allBids, seatBid.Bid...)
			}
		}
	}

	// Check for errors
	for err := range errors {
		log.Printf("Bid request error: %v", err)
	}

	if len(allBids) == 0 {
		return &BidResponse{
			ID:  bidRequest.ID,
			NBR: 2, // No bid
		}, nil
	}

	// Determine winning bid
	winner := e.selectWinningBid(allBids, bidRequest.At)

	// Record auction result
	e.recordAuctionResult(ctx, bidRequest.ID, winner, allBids)

	return &BidResponse{
		ID:      bidRequest.ID,
		SeatBid: []SeatBid{{Bid: []Bid{*winner}}},
		Cur:     "USD",
	}, nil
}

// sendBidRequest sends a bid request to a single bidder
func (e *RTBEngine) sendBidRequest(ctx context.Context, bidder Bidder, bidRequest BidRequest) (*BidResponse, error) {
	client := &http.Client{Timeout: e.bidTimeout}

	reqBody, err := json.Marshal(bidRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bid request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", bidder.Endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bidder.AuthToken)
	req.Header.Set("x-openrtb-version", "2.5")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bid request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bid request returned status %d", resp.StatusCode)
	}

	var bidResponse BidResponse
	if err := json.NewDecoder(resp.Body).Decode(&bidResponse); err != nil {
		return nil, fmt.Errorf("failed to decode bid response: %w", err)
	}

	return &bidResponse, nil
}

// selectWinningBid selects the winning bid based on auction type
func (e *RTBEngine) selectWinningBid(bids []Bid, auctionType int) *Bid {
	if len(bids) == 0 {
		return nil
	}

	// Sort bids by price (descending)
	for i := 0; i < len(bids)-1; i++ {
		for j := i + 1; j < len(bids); j++ {
			if bids[j].Price > bids[i].Price {
				bids[i], bids[j] = bids[j], bids[i]
			}
		}
	}

	winner := bids[0]

	// Second-price auction adjustment
	if auctionType == 2 && len(bids) > 1 {
		secondPrice := bids[1].Price
		winner.Price = secondPrice + 0.01 // $0.01 above second price
	}

	return &winner
}

// recordAuctionResult records the auction result for analytics
func (e *RTBEngine) recordAuctionResult(ctx context.Context, requestID string, winner *Bid, allBids []Bid) {
	result := map[string]interface{}{
		"request_id": requestID,
		"winner_id":  winner.ID,
		"win_price":  winner.Price,
		"bid_count":  len(allBids),
		"timestamp":  time.Now().Unix(),
	}

	data, _ := json.Marshal(result)
	e.redis.Publish(ctx, "rtb:auctions", string(data))
}

// GetBidderStats returns statistics for a bidder
func (e *RTBEngine) GetBidderStats(bidderID string) (map[string]interface{}, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	bidder, exists := e.bidders[bidderID]
	if !exists {
		return nil, fmt.Errorf("bidder not found")
	}

	return map[string]interface{}{
		"id":           bidder.ID,
		"name":         bidder.Name,
		"enabled":      bidder.Enabled,
		"avg_response": bidder.AvgResponse.Milliseconds(),
		"win_rate":     bidder.WinRate,
	}, nil
}

// UpdateBidderPerformance updates bidder performance metrics
func (e *RTBEngine) UpdateBidderPerformance(bidderID string, responseTime time.Duration, won bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if bidder, exists := e.bidders[bidderID]; exists {
		// Exponential moving average for response time
		alpha := 0.1
		bidder.AvgResponse = time.Duration(float64(bidder.AvgResponse)*(1-alpha) + float64(responseTime)*alpha)

		// Update win rate
		if won {
			bidder.WinRate = bidder.WinRate*0.9 + 0.1
		} else {
			bidder.WinRate = bidder.WinRate * 0.9
		}
	}
}
