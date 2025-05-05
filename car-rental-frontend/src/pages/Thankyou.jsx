import React from "react";
import "./Thankyou.css";
import jeen from "../assets/jeen.png";
import tos from "../assets/tos.png";
import bb  from "../assets/bb.png";
import win from "../assets/win.png";
import neww from "../assets/new.png";
import manyl from "../assets/manyl.png";

const people = [
  {
    name: "KANTAWICH LIMWILAI",
    id: "66070506406",
    image: bb,
  },
  {
    name: "CHANNATHAT UEANAPAPHON",
    id: "66070506413",
    image: tos,
  },
  {
    name: "THAPANA LIAMTHONGKAOW",
    id: "66070506416",
    image: win,
  },
  {
    name: "SIRAWAT SANBOONSONG",
    id: "66070506456",
    image: neww,
  },
  {
    name: "ACHITA CHEARDSATIENSAK",
    id: "66070506486",
    image: jeen,
  },
  {
    name: "MANYL HAMDI AISSA",
    id: "67540460065",
    image: manyl,
  },
];

const PresentedBy = () => {
  return (
    <div className="presented-container">
      <h1 className="title">Presented By</h1>
      <div className="grid">
        {people.map((person, index) => (
          <div key={index} className="card">
            <img src={person.image} alt={person.name} className="image" />
            <p className="name">{person.name}</p>
            <p className="id">{person.id}</p>
          </div>
        ))}
      </div>
    </div>
  );
};

export default PresentedBy;
