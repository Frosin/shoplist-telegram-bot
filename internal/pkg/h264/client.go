package h264

import (
	"bytes"
	"errors"
	"image/jpeg"
	"log"
	"strconv"
	"time"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/bluenviron/gortsplib/v4/pkg/format/rtph264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pion/rtp"
)

func GetStream(url string, delay time.Duration, count int) chan *tgbotapi.FileBytes {
	outChan := make(chan *tgbotapi.FileBytes)

	go func() {
		defer close(outChan)

		ticker := time.NewTicker(delay)
		defer ticker.Stop()

		c := gortsplib.Client{}

		u, err := base.ParseURL(url)
		if err != nil {
			log.Printf("ERR: ParseURL: %v", err)
			return
		}

		err = c.Start(u.Scheme, u.Host)
		if err != nil {
			log.Printf("ERR: Start: %v", err)
			return
		}
		defer c.Close()

		// find available medias
		desc, _, err := c.Describe(u)
		if err != nil {
			log.Printf("ERR: describe: %v", err)
			return
		}

		// find the H264 media and format
		var forma *format.H264
		medi := desc.FindFormat(&forma)
		if medi == nil {
			log.Printf("ERR: FindFormat: %v", errors.New("media not found"))
			return
		}

		// setup RTP -> H264 decoder
		rtpDec, err := forma.CreateDecoder()
		if err != nil {
			log.Printf("ERR: CreateDecoder: %v", errors.New("media not found"))
			return
		}

		// setup H264 -> raw frames decoder
		frameDec := &h264Decoder{}
		err = frameDec.initialize()
		if err != nil {
			log.Printf("ERR: initialize: %v", errors.New("media not found"))
			return
		}
		defer frameDec.close()

		// if SPS and PPS are present into the SDP, send them to the decoder
		if forma.SPS != nil {
			frameDec.decode(forma.SPS)
		}
		if forma.PPS != nil {
			frameDec.decode(forma.PPS)
		}

		// setup a single media
		_, err = c.Setup(desc.BaseURL, medi, 0, 0)
		if err != nil {
			log.Printf("ERR: Setup: %v", errors.New("media not found"))
			return
		}

		iframeReceived := false
		saveCount := 0

		// called when a RTP packet arrives
		c.OnPacketRTP(medi, forma, func(pkt *rtp.Packet) {
			// extract access units from RTP packets
			au, err := rtpDec.Decode(pkt)
			if err != nil {
				if err != rtph264.ErrNonStartingPacketAndNoPrevious && err != rtph264.ErrMorePacketsNeeded {
					log.Printf("ERR: Decode: %v", err)
				}
				return
			}

			// wait for an I-frame
			if !iframeReceived {
				if !h264.IsRandomAccess(au) {
					log.Printf("waiting for an I-frame")
					return
				}
				iframeReceived = true
			}

			for _, nalu := range au {
				// convert NALUs into RGBA frames
				img, err := frameDec.decode(nalu)
				if err != nil {
					log.Println(err)
					return
				}

				// wait for a frame
				if img == nil {
					continue
				}

				buf := new(bytes.Buffer)
				err = jpeg.Encode(buf, img, &jpeg.Options{
					Quality: 60,
				})
				if err != nil {
					log.Println(err)
					return
				}

				select {
				case <-ticker.C:
					outChan <- &tgbotapi.FileBytes{Name: strconv.FormatInt(time.Now().UnixMilli(), 10), Bytes: buf.Bytes()}
					saveCount++
				default:
				}

				break
			}

			if saveCount == count {
				c.Close()
				return
			}

		})

		// start playing
		_, err = c.Play(nil)
		if err != nil {
			log.Println(err)
			return
		}

		// wait until a fatal error
		log.Println(c.Wait())
	}()

	return outChan
}
