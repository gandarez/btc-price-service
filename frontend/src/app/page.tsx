'use client';

import { useEffect, useRef, useState } from 'react';
import Image from 'next/image';

export default function Home() {
  const [price, setPrice] = useState<number | null>(null);
  const [timestamp, setTimestamp] = useState<string | null>(null);
  const [status, setStatus] = useState("Connecting...");
  const [priceClass, setPriceClass] = useState(""); // for color + animation
  const prevPriceRef = useRef<number | null>(null);
  const retryInterval = useRef<NodeJS.Timeout | null>(null);
  const eventSourceRef = useRef<EventSource | null>(null);

  const formatter = new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });

  useEffect(() => {
    const connect = () => {
    setStatus("Connecting...");
    const source = new EventSource("/api/stream");
    eventSourceRef.current = source;

    source.onopen = () => {
      setStatus("Connected ✅");
      if (retryInterval.current) {
        clearInterval(retryInterval.current);
        retryInterval.current = null;
      }
    };

    source.onerror = () => {
      setStatus("Disconnected 🔌 Retrying in 2s...");
      source.close();

      // Retry connection every 2 seconds if not already trying
      if (!retryInterval.current) {
        retryInterval.current = setInterval(() => {
          connect();
        }, 2000);
      }
    };

    source.onmessage = (event) => {
      try {
        if (event.data === "ping" || event.data === "connected") {
          // Ignore ping messages
          return;
        }

        const data = JSON.parse(event.data);
        const newPrice = Number(data.price);

        if (!isNaN(newPrice)) {
          const prev = prevPriceRef.current;
          let directionClass = "";

          if (prev !== null) {
            directionClass =
              newPrice > prev ? "price-up" :
              newPrice < prev ? "price-down" : "";
          }

          setPriceClass(directionClass);
          prevPriceRef.current = newPrice;
          setPrice(newPrice);
          setTimestamp(new Date(data.timestamp).toLocaleString());

          // Remove animation class after it plays
          setTimeout(() => setPriceClass(""), 600);
        }
      } catch (err) {
        console.error("Parse error:", err);
      }
    };
  };

  connect();

  // Clean up on unmount
  return () => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }
    if (retryInterval.current) {
      clearInterval(retryInterval.current);
    }
  };
}, []);

  return (
    <main style={styles.container}>
      <Image src="/logo.png" alt="Injective Labs Logo" width={120} height={120} style={styles.logo} />
      <div style={styles.priceBox}>
        <h1 style={styles.header}>Live BTC/USD Price</h1>
        <div className={priceClass} style={styles.price}>
          {price !== null ? formatter.format(price) : "--"}
        </div>
        <div style={styles.timestamp}>{timestamp ? `Updated at: ${timestamp}` : "Waiting for update..."}</div>
        <div style={styles.status}>{status}</div>
      </div>
    </main>
  );
}

const styles: { [key: string]: React.CSSProperties } = {
  container: {
    backgroundColor: "#14151a",
    color: "#e0e0e0",
    minHeight: "100vh",
    display: "flex",
    flexDirection: "column",
    justifyContent: "center",
    alignItems: "center",
    fontFamily: "Segoe UI, sans-serif",
    padding: "2rem",
    textAlign: "center"
  },
  logo: {
    width: "120px",
    height: "auto",
    marginBottom: "2rem",
  },
  priceBox: {
    backgroundColor: "#121212",
    border: "1px solid #333",
    borderRadius: "12px",
    padding: "2rem",
    boxShadow: "0 4px 8px rgba(0, 0, 0, 0.3)",
    maxWidth: "400px",
    width: "100%"
  },
  header: {
    color: "#f7931a",
    fontSize: "2rem",
    margin: "0 0 1rem 0"
  },
  price: {
    fontSize: "3rem",
    margin: "1rem 0"
  },
  timestamp: {
    fontSize: "1rem",
    color: "#aaa"
  },
  status: {
    marginTop: "1rem",
    fontSize: "0.9rem"
  }
};

<style jsx>{`
  .price-up {
    color: #4caf50;
    animation: flash 0.6s ease;
  }

  .price-down {
    color: #f44336;
    animation: flash 0.6s ease;
  }

  @keyframes flash {
    0% { background-color: rgba(255,255,255,0.1); }
    100% { background-color: transparent; }
  }
`}</style>