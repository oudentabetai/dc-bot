package gifmaker

import (
	"fmt"
	"io"
	"os"
	"log"
	"net/http"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func ConvertToGif(attachmentURL string) (error) {
	
	res, err := http.DefaultClient.Get(attachmentURL)

	defer res.Body.Close()

	filePath := "./in.png"
	out, err := os.Create(filePath)
	if err != nil {
	    log.Printf("Failed to create file: %v", err)
	    return err
	}
	defer out.Close()

	// 2. HTTPレスポンスのBodyをファイルにコピー（書き込み）する
	written, err := io.Copy(out, res.Body)
	if err != nil {
	    log.Printf("Failed to save image: %v", err)
	    return err
	}

	log.Printf("Successfully saved %d bytes to %s", written, filePath)



	// 一時ファイルのパスを設定
	palettePath := "./palette.png"
	tmpOutput := "./out.gif"

	// 処理完了後にパレット用の一時画像は確実に削除する
	defer os.Remove(palettePath)

	// 1. palettegen パス: 最適なパレット画像(256色)を生成
	fmt.Println("[1/2] 最適なカラーパレットを生成中...")
	err = ffmpeg.Input(filePath).
		Output(palettePath, ffmpeg.KwArgs{"vf": "palettegen"}).
		OverWriteOutput(). 
		Run()
	if err != nil {
		return fmt.Errorf("palettegen エラー: %w", err)
	}

	// 2. paletteuse パス: 生成したパレットを使って最高画質でGIF化
	fmt.Println("[2/2] パレットを適用してGIFを出力中...")
	err = ffmpeg.Filter(
		[]*ffmpeg.Stream{
			ffmpeg.Input(filePath),
			ffmpeg.Input(palettePath),
		},
		"paletteuse", 
		ffmpeg.Args{},
	).
		Output(tmpOutput, ffmpeg.KwArgs{"loop": "0"}).
		OverWriteOutput().
		Run()

	if err != nil {
		return fmt.Errorf("paletteuse エラー: %w", err)
	}

	return nil
}

// ファイルをコピーするヘルパー関数
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	return nil
}